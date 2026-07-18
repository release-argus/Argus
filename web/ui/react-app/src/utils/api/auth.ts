import type {
	APIToken,
	APITokenCreated,
	AuthGroup,
	AuthMe,
	AuthUser,
	PermissionCatalogue,
} from '@/types/auth';
import { API_BASE } from '@/utils/api/types/api-request';
import type {
	APITokenCreateRequest,
	GroupCreateRequest,
	GroupPatchRequest,
	LoginRequest,
	SetupRequest,
	SetupState,
	UserCreateRequest,
	UserPatchRequest,
} from '@/utils/api/types/requests/auth';
import fetchJSON from '@/utils/fetch-json';

/**
 * Auth/RBAC API calls. The backend is the single authority on permissions;
 * these calls only fetch/act - the UI never re-evaluates access rules.
 */

export const login = (credentials: LoginRequest) =>
	fetchJSON<AuthMe>({
		body: JSON.stringify(credentials),
		method: 'POST',
		url: `${API_BASE}/auth/login`,
	});

export const logout = () =>
	fetchJSON<null>({
		method: 'POST',
		url: `${API_BASE}/auth/logout`,
	});

export const fetchMe = () =>
	fetchJSON<AuthMe>({
		url: `${API_BASE}/auth/me`,
	});

// First-run setup (unauthenticated; the backend 409s once any user exists).
export const fetchSetupState = () =>
	fetchJSON<SetupState>({ url: `${API_BASE}/auth/setup` });

export const setup = (account: SetupRequest) =>
	fetchJSON<AuthMe>({
		body: JSON.stringify(account),
		method: 'POST',
		url: `${API_BASE}/auth/setup`,
	});

// Users.
export const listUsers = () =>
	fetchJSON<AuthUser[]>({ url: `${API_BASE}/users` });

export const createUser = (user: UserCreateRequest) =>
	fetchJSON<AuthUser>({
		body: JSON.stringify(user),
		method: 'POST',
		url: `${API_BASE}/users`,
	});

export const updateUser = (id: string, patch: UserPatchRequest) =>
	fetchJSON<AuthUser>({
		body: JSON.stringify(patch),
		method: 'PATCH',
		url: `${API_BASE}/users/${encodeURIComponent(id)}`,
	});

export const deleteUser = (id: string) =>
	fetchJSON<null>({
		method: 'DELETE',
		url: `${API_BASE}/users/${encodeURIComponent(id)}`,
	});

// Groups.
export const listGroups = () =>
	fetchJSON<AuthGroup[]>({ url: `${API_BASE}/groups` });

export const createGroup = (group: GroupCreateRequest) =>
	fetchJSON<AuthGroup>({
		body: JSON.stringify(group),
		method: 'POST',
		url: `${API_BASE}/groups`,
	});

export const updateGroup = (id: string, patch: GroupPatchRequest) =>
	fetchJSON<AuthGroup>({
		body: JSON.stringify(patch),
		method: 'PATCH',
		url: `${API_BASE}/groups/${encodeURIComponent(id)}`,
	});

export const deleteGroup = (id: string) =>
	fetchJSON<null>({
		method: 'DELETE',
		url: `${API_BASE}/groups/${encodeURIComponent(id)}`,
	});

// Permission catalogue (read-only; grants are edited on groups).
export const fetchPermissionCatalogue = () =>
	fetchJSON<PermissionCatalogue>({ url: `${API_BASE}/permissions` });

// API tokens (own-scoped).
export const listTokens = () =>
	fetchJSON<APIToken[]>({ url: `${API_BASE}/tokens` });

export const createToken = (token: APITokenCreateRequest) =>
	fetchJSON<APITokenCreated>({
		body: JSON.stringify(token),
		method: 'POST',
		url: `${API_BASE}/tokens`,
	});

export const deleteToken = (id: string) =>
	fetchJSON<null>({
		method: 'DELETE',
		url: `${API_BASE}/tokens/${encodeURIComponent(id)}`,
	});
