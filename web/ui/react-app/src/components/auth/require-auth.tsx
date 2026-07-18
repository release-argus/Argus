import { LoaderCircle } from 'lucide-react';
import type { ReactElement, ReactNode } from 'react';
import { Navigate, useLocation } from 'react-router';
import { useAuth } from '@/contexts/auth';

type RequireAuthProps = {
	children: ReactNode;
};

/**
 * Gates the app shell behind authentication. Unauthenticated visitors (first
 * visit, logged out, or an expired session) are sent to /login, remembering
 * where they were so a successful login returns them there.
 */
export const RequireAuth = (props: RequireAuthProps): ReactElement => {
	const { status } = useAuth();
	const location = useLocation();

	switch (status) {
		case 'loading':
			return (
				<output
					aria-label="Checking authentication"
					className="flex min-h-[50vh] w-full items-center justify-center"
				>
					<LoaderCircle aria-hidden className="size-8 animate-spin" />
				</output>
			);
		case 'unauthenticated':
		case 'setup':
			return (
				<Navigate
					replace
					state={{ from: location.pathname + location.search }}
					to="/login"
				/>
			);
		default:
			return <>{props.children}</>;
	}
};
