import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
	createContext,
	type ReactNode,
	use,
	useCallback,
	useEffect,
	useMemo,
	useRef,
	useState,
} from 'react';
import { disconnectWebSocket } from '@/contexts/websocket';
import { UNAUTHORISED_EVENT } from '@/lib/auth-events';
import { QUERY_KEYS } from '@/lib/query-keys';
import {
	type Action,
	ADMIN_GROUP,
	type AuthMe,
	type Grant,
	type Resource,
} from '@/types/auth';
import * as authAPI from '@/utils/api/auth';
import { APIError } from '@/utils/errors';

type AuthStatus =
	/** Initial /auth/me fetch in flight. */
	| 'loading'
	/** Auth is off (not a disabled account) - everything is allowed, no login exists. */
	| 'auth-disabled'
	/** Auth is on and no users exist yet - first-run setup is pending. */
	| 'setup'
	/** Auth is on and nobody is logged in. */
	| 'unauthenticated'
	/** Auth is on and a user is logged in. */
	| 'authenticated';

type PermissionTarget = {
	/** ID of the targeted service (for service-scoped grants). */
	serviceID?: string;
	/** Dashboard tags of the targeted service (for service_tag grants). */
	tags?: string[];
};

type AuthContextProps = {
	status: AuthStatus;
	user?: AuthMe['user'];
	permissions: Grant[];
	/** Whether the user belongs to the admin group. */
	isAdmin: boolean;
	/**
	 * Whether the user holds (resource, action) for the optional target.
	 * Mirrors the grant list from /auth/me.
	 */
	hasPermission: (
		resource: Resource,
		action: Action,
		target?: PermissionTarget,
	) => boolean;
	/**
	 * Whether the user holds (resource, action) under ANY scope - e.g. to
	 * decide whether edit permissions exist at all, before per-target checks.
	 */
	hasAnyPermission: (resource: Resource, action: Action) => boolean;
	login: (username: string, password: string) => Promise<void>;
	logout: () => Promise<void>;
	/** Creates the first administrator (first-run setup) and logs them in. */
	setup: (
		username: string,
		displayName: string,
		password: string,
	) => Promise<void>;
};

export const AuthContext = createContext<AuthContextProps>({
	hasAnyPermission: () => false,
	hasPermission: () => false,
	isAdmin: false,
	login: () => Promise.resolve(),
	logout: () => Promise.resolve(),
	permissions: [],
	setup: () => Promise.resolve(),
	status: 'loading',
});

type AuthProviderProps = {
	children: ReactNode;
};

/**
 * Resolves and holds the authenticated user and their permission grants.
 *
 * A 404 from /auth/me means auth is disabled; a 401 means login is required.
 */
export const AuthProvider = (props: AuthProviderProps) => {
	const queryClient = useQueryClient();
	const [status, setStatus] = useState<AuthStatus>('loading');

	const { data: me, error: meError } = useQuery({
		enabled: status !== 'auth-disabled',
		queryFn: authAPI.fetchMe,
		queryKey: QUERY_KEYS.AUTH.ME(),
		retry: false,
	});

	// The user /auth/me last resolved to, so a change of identity can be seen.
	const lastUserID = useRef<string | undefined>(undefined);

	const applyMe = useCallback(
		(data: AuthMe) => {
			// Drop any prior session's cached data on login.
			queryClient.clear();
			queryClient.setQueryData(QUERY_KEYS.AUTH.ME(), data);
			lastUserID.current = data.user.id;
			setStatus('authenticated');
		},
		[queryClient],
	);

	// Resolve the auth status from the /auth/me query.
	useEffect(() => {
		if (meError) {
			if (meError instanceof APIError && meError.status === 404) {
				setStatus('auth-disabled');
				return;
			}
			// A refetch failure that is not a 401 (e.g. a network blip) keeps
			// the session - the last-known `me` still stands.
			if (!(meError instanceof APIError && meError.status === 401) && me)
				return;
			// 401 (or anything else) requires a login to proceed - unless
			// no users exist yet, in which case first-run setup comes first.
			let cancelled = false;
			authAPI
				.fetchSetupState()
				.then(({ setup_required }) => {
					if (!cancelled)
						setStatus(setup_required ? 'setup' : 'unauthenticated');
				})
				.catch(() => {
					if (!cancelled) setStatus('unauthenticated');
				});
			return () => {
				cancelled = true;
			};
		}
		if (me) {
			// A different user answering /auth/me means the cache belongs to the
			// previous session (e.g. signed in elsewhere after this one expired).
			if (
				lastUserID.current !== undefined &&
				lastUserID.current !== me.user.id
			) {
				queryClient.clear();
			}
			lastUserID.current = me.user.id;
			setStatus('authenticated');
		}
	}, [me, meError, queryClient]);

	// A 401 anywhere (e.g. expired session mid-use) drops back to the login
	// page (setup stays put). The cache and the socket both belong to the
	// dead session, so they go too.
	useEffect(() => {
		const onUnauthorised = () => {
			if (status === 'auth-disabled' || status === 'setup') return;
			disconnectWebSocket();
			queryClient.clear();
			setStatus('unauthenticated');
		};
		globalThis.addEventListener(UNAUTHORISED_EVENT, onUnauthorised);
		return () =>
			globalThis.removeEventListener(UNAUTHORISED_EVENT, onUnauthorised);
	}, [queryClient, status]);

	const login = useCallback(
		async (username: string, password: string) => {
			const data = await authAPI.login({ password, username });
			applyMe(data);
		},
		[applyMe],
	);

	const setup = useCallback(
		async (username: string, displayName: string, password: string) => {
			const data = await authAPI.setup({
				display_name: displayName || undefined,
				password,
				username,
			});
			applyMe(data);
		},
		[applyMe],
	);

	const logout = useCallback(async () => {
		try {
			await authAPI.logout();
		} finally {
			disconnectWebSocket();
			queryClient.clear();
			setStatus('unauthenticated');
		}
	}, [queryClient]);

	const permissions = useMemo(() => me?.permissions ?? [], [me]);

	const isAdmin = useMemo(
		() => me?.user?.groups?.includes(ADMIN_GROUP) ?? false,
		[me],
	);

	const hasPermission = useCallback(
		(
			resource: Resource,
			action: Action,
			target?: PermissionTarget,
		): boolean => {
			// No auth = no restrictions.
			if (status === 'auth-disabled') return true;

			return permissions.some((grant) => {
				if (grant.resource !== resource || grant.action !== action)
					return false;
				if (grant.scope.type !== 'global' && !target?.serviceID) return false;
				switch (grant.scope.type) {
					case 'global':
						return true;
					case 'service':
						return grant.scope.ref === target?.serviceID;
					case 'service_tag':
						return (
							grant.scope.ref !== undefined &&
							(target?.tags ?? []).includes(grant.scope.ref)
						);
					default:
						return false;
				}
			});
		},
		[permissions, status],
	);

	const hasAnyPermission = useCallback(
		(resource: Resource, action: Action): boolean => {
			if (status === 'auth-disabled') return true;
			return permissions.some(
				(grant) => grant.resource === resource && grant.action === action,
			);
		},
		[permissions, status],
	);

	const contextValue = useMemo(
		() => ({
			hasAnyPermission,
			hasPermission,
			isAdmin,
			login,
			logout,
			permissions,
			setup,
			status,
			user: me?.user,
		}),
		[
			hasAnyPermission,
			hasPermission,
			isAdmin,
			login,
			logout,
			permissions,
			setup,
			status,
			me,
		],
	);

	return <AuthContext value={contextValue}>{props.children}</AuthContext>;
};

/**
 * @returns The auth context: status, user, permissions, and actions.
 */
export const useAuth = () => {
	return use(AuthContext);
};
