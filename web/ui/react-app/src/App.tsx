import {
	MutationCache,
	QueryCache,
	QueryClient,
	QueryClientProvider,
} from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import type { ReactElement } from 'react';
import {
	Navigate,
	Outlet,
	Route,
	BrowserRouter as Router,
	Routes,
} from 'react-router';
import { RequireAuth } from '@/components/auth/require-auth';
import Header from '@/components/header';
import { ThemeProvider } from '@/components/theme-provider';
import { Toaster } from '@/components/ui/sonner';
import { AuthProvider, notifyUnauthorised, useAuth } from '@/contexts/auth';
import { ModalProvider } from '@/contexts/modal';
import { WebSocketProvider } from '@/contexts/websocket';
import { QUERY_KEYS } from '@/lib/query-keys';
import {
	ApprovalsPage,
	ConfigPage,
	FlagsPage,
	GroupsPage,
	LoginPage,
	StatusPage,
	TokensPage,
	UsersPage,
} from '@/pages';
import { getBasename } from '@/utils';
import { APIError } from '@/utils/errors';

/**
 * Any 401 from the API (e.g. a session that expired mid-use) drops the app
 * back to the login page via the auth context.
 */
const onAPIError = (error: unknown) => {
	if (error instanceof APIError && error.status === 401) notifyUnauthorised();
};

/**
 * /auth/me resolves its own 401s in the auth context (deciding between the
 * login and first-run setup pages), so it stays out of the global handler.
 */
const ME_QUERY_KEY = QUERY_KEYS.AUTH.ME().join('/');

/**
 * The protected app shell: everything except /login sits behind
 * authentication (when auth is enabled).
 */
const ProtectedApp = (): ReactElement => {
	const { isAdmin } = useAuth();

	return (
		<RequireAuth>
			<Header />
			<WebSocketProvider>
				<Toaster expand richColors visibleToasts={3} />
				<ModalProvider>
					<div className="w-full p-5">
						<Routes>
							<Route element={<Navigate to="/approvals" />} path="/" />
							<Route element={<ApprovalsPage />} path="/approvals" />
							<Route element={<StatusPage />} path="/status" />
							<Route element={<FlagsPage />} path="/flags" />
							<Route element={<ConfigPage />} path="/config" />
							<Route
								element={
									isAdmin ? <Outlet /> : <Navigate replace to="/approvals" />
								}
								path="/admin"
							>
								<Route element={<Navigate replace to="users" />} index />
								<Route element={<UsersPage />} path="users" />
								<Route element={<GroupsPage />} path="groups" />
							</Route>
							<Route element={<Outlet />} path="/account">
								<Route element={<Navigate replace to="tokens" />} index />
								<Route element={<TokensPage />} path="tokens" />
							</Route>
						</Routes>
					</div>
				</ModalProvider>
			</WebSocketProvider>
		</RequireAuth>
	);
};

const App = (): ReactElement => {
	// Determine `pathPrefix` by stripping the first known endpoint suffix from the window location path.
	const basename = getBasename();

	const queryClient = new QueryClient({
		mutationCache: new MutationCache({ onError: onAPIError }),
		queryCache: new QueryCache({
			onError: (error, query) => {
				if (query.queryKey.join('/') === ME_QUERY_KEY) return;
				onAPIError(error);
			},
		}),
	});
	queryClient.setDefaultOptions({
		queries: {
			gcTime: 1000 * 60 * 10, // 10 minutes.
			refetchOnWindowFocus: true,
			staleTime: 1000 * 60 * 5, // 5 minutes.
		},
	});

	return (
		<ThemeProvider>
			<QueryClientProvider client={queryClient}>
				<Router basename={basename}>
					<AuthProvider>
						<Routes>
							<Route element={<LoginPage />} path="/login" />
							<Route element={<ProtectedApp />} path="*" />
						</Routes>
					</AuthProvider>
				</Router>
				<ReactQueryDevtools initialIsOpen={false} />
			</QueryClientProvider>
		</ThemeProvider>
	);
};

export default App;
