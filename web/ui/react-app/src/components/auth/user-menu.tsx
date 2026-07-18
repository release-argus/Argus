import { KeyRound, LogOut, UserCircle, Users } from 'lucide-react';
import type { ReactElement } from 'react';
import { Link, useNavigate } from 'react-router';
import { Button } from '@/components/ui/button';
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useAuth } from '@/contexts/auth';

/**
 * The signed-in user's menu: token management, admin pages the user may see,
 * and logout. Renders nothing when auth is disabled or nobody is logged in.
 */
export const UserMenu = (): ReactElement | null => {
	const { status, user, isAdmin, logout } = useAuth();
	const navigate = useNavigate();

	if (status !== 'authenticated' || !user) return null;

	const onLogout = () => {
		void logout()
			.catch(() => {
				// logout() clears the local session even if the API call fails,
				// so ignore its rejection to avoid an unhandled promise rejection.
			})
			.finally(() => navigate('/login', { replace: true }));
	};

	return (
		<DropdownMenu>
			<DropdownMenuTrigger asChild>
				<Button aria-label="User menu" size="icon-md" variant="outline">
					<UserCircle aria-hidden />
				</Button>
			</DropdownMenuTrigger>
			<DropdownMenuContent align="end" className="z-100 min-w-48">
				<DropdownMenuLabel className="truncate">
					{user.display_name || user.username}
				</DropdownMenuLabel>
				<DropdownMenuSeparator />
				<DropdownMenuItem asChild>
					<Link to="/account/tokens">
						<KeyRound aria-hidden /> API Tokens
					</Link>
				</DropdownMenuItem>
				{isAdmin && (
					<>
						<DropdownMenuItem asChild>
							<Link to="/admin/users">
								<Users aria-hidden /> Users
							</Link>
						</DropdownMenuItem>
						<DropdownMenuItem asChild>
							<Link to="/admin/groups">
								<Users aria-hidden /> Groups
							</Link>
						</DropdownMenuItem>
					</>
				)}
				<DropdownMenuSeparator />
				<DropdownMenuItem onClick={onLogout}>
					<LogOut aria-hidden /> Log out
				</DropdownMenuItem>
			</DropdownMenuContent>
		</DropdownMenu>
	);
};
