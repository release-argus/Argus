import { ApprovalsPage, ConfigPage, FlagsPage, StatusPage } from "pages";
import {
  Navigate,
  Route,
  BrowserRouter as Router,
  Routes,
} from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { Container } from "react-bootstrap";
import Header from "components/header";
import { ModalProvider } from "contexts/modal";
import { NotificationProvider } from "contexts/notification";
import { ReactElement } from "react";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { WebSocketProvider } from "contexts/websocket";
import { getBasename } from "utils";

const App = (): ReactElement => {
  // This dynamically/generically determines the pathPrefix by stripping the first known
  // endpoint suffix from the window location path. It works out of the box for both direct
  // hosting and reverse proxy deployments with no additional configurations required.
  const basename = getBasename();

  const queryClient = new QueryClient();
  queryClient.setDefaultOptions({
    queries: {
      gcTime: 1000 * 60 * 10, // 10 minutes
      refetchOnWindowFocus: true,
      staleTime: 1000 * 60 * 5, // 5 minutes
    },
  });

  return (
    <QueryClientProvider client={queryClient}>
      <Router basename={basename}>
        <Header />
        <WebSocketProvider>
          <NotificationProvider />
          <ModalProvider>
            <Container fluid style={{ padding: "1.25rem" }}>
              <Routes>
                <Route path="/approvals" element={<ApprovalsPage />} />
                <Route path="/status" element={<StatusPage />} />
                <Route path="/flags" element={<FlagsPage />} />
                <Route path="/config" element={<ConfigPage />} />
                <Route path="/" element={<Navigate to="/approvals" />} />
              </Routes>
            </Container>
          </ModalProvider>
        </WebSocketProvider>
      </Router>
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
};

export default App;
