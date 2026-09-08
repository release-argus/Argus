import { ChevronDown, KeyRound, UserCircle } from 'lucide-react';
import type { ReactElement } from 'react';
import { Link, NavLink, Outlet, useLocation } from 'react-router';
import { Button } from '@/components/ui/button';
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { cn } from '@/lib/utils';

const SECTIONS = [
	{ icon: UserCircle, label: 'Account', to: '/account/profile' },
	{ icon: KeyRound, label: 'API Tokens', to: '/account/tokens' },
];

const linkStyle = (isActive: boolean) =>
	cn(
		'flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors',
		isActive
			? 'bg-muted font-medium text-foreground'
			: 'text-muted-foreground hover:bg-muted/50 hover:text-foreground',
	);

/**
 * The settings shell: its sections as a sidebar on wide screens, and as a
 * dropdown below that.
 */
export const SettingsLayout = (): ReactElement => {
	const { pathname } = useLocation();
	const active =
		SECTIONS.find((section) => pathname.startsWith(section.to)) ?? SECTIONS[0];
	const ActiveIcon = active.icon;

	return (
		<div className="flex flex-col gap-4">
			<h2 className="scroll-m-20 font-semibold text-3xl tracking-tight">
				Settings
			</h2>
			<div className="flex flex-col gap-6 lg:flex-row lg:gap-10">
				<nav
					aria-label="Settings sections"
					className="hidden lg:block lg:w-52 lg:shrink-0"
				>
					<ul className="flex flex-col gap-1">
						{SECTIONS.map(({ icon: Icon, label, to }) => (
							<li key={to}>
								<NavLink
									className={({ isActive }) => linkStyle(isActive)}
									to={to}
								>
									<Icon aria-hidden className="size-4" />
									{label}
								</NavLink>
							</li>
						))}
					</ul>
				</nav>
				<div className="lg:hidden">
					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<Button
								aria-label="Settings sections"
								className="w-full justify-between"
								variant="outline"
							>
								<span className="flex items-center gap-2">
									<ActiveIcon aria-hidden />
									{active.label}
								</span>
								<ChevronDown aria-hidden />
							</Button>
						</DropdownMenuTrigger>
						<DropdownMenuContent
							align="start"
							className="w-(--radix-dropdown-menu-trigger-width)"
						>
							{SECTIONS.map(({ icon: Icon, label, to }) => (
								<DropdownMenuItem asChild key={to}>
									<Link to={to}>
										<Icon aria-hidden />
										{label}
									</Link>
								</DropdownMenuItem>
							))}
						</DropdownMenuContent>
					</DropdownMenu>
				</div>
				<div className="min-w-0 flex-1">
					<Outlet />
				</div>
			</div>
		</div>
	);
};
