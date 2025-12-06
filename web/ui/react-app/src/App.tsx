import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import type { ReactElement } from 'react';
import {
	Navigate,
	Route,
	BrowserRouter as Router,
	Routes,
} from 'react-router-dom';
import Header from '@/components/header';
import { ThemeProvider } from '@/components/theme-provider';
import { Toaster } from '@/components/ui/sonner';
import { ModalProvider } from '@/contexts/modal';
import { WebSocketProvider } from '@/contexts/websocket';
import { ApprovalsPage, ConfigPage, FlagsPage, StatusPage } from '@/pages';
import { getBasename } from '@/utils';

const App = (): ReactElement => {
	// Determine `pathPrefix` by stripping the first known endpoint suffix from the window location path.
	// Works for both direct hosting and reverse proxy deployments.
	const basename = getBasename();

	const queryClient = new QueryClient();
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
					<Header />
					<WebSocketProvider>
						<Toaster expand richColors visibleToasts={3} />
						<ModalProvider>
							<div className="w-full p-5">
								<Routes>
									<Route element={<ApprovalsPage />} path="/approvals" />
									<Route element={<StatusPage />} path="/status" />
									<Route element={<FlagsPage />} path="/flags" />
									<Route element={<ConfigPage />} path="/config" />
									<Route element={<Navigate to="/approvals" />} path="/" />
								</Routes>
							</div>
						</ModalProvider>
					</WebSocketProvider>
				</Router>
				<ReactQueryDevtools initialIsOpen={false} />
			</QueryClientProvider>
		</ThemeProvider>
	);
};

export default App;
