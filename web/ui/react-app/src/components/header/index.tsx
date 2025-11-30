import { ChevronDown, Menu } from 'lucide-react';
import { useState } from 'react';
import { Link } from 'react-router-dom';
import { ThemeModeToggle } from '@/components/theme-mode-toggle';
import { Button } from '@/components/ui/button';
import {
	Collapsible,
	CollapsibleContent,
	CollapsibleTrigger,
} from '@/components/ui/collapsible';
import {
	NavigationMenu,
	NavigationMenuContent,
	NavigationMenuItem,
	NavigationMenuLink,
	NavigationMenuList,
	NavigationMenuTrigger,
} from '@/components/ui/navigation-menu';
import { cn } from '@/lib/utils';

const ArgusGitHubRepo = 'https://github.com/release-argus/Argus';
const headerOptions = [
	{
		href: '/approvals',
		name: 'Approvals',
	},
	{
		children: [
			{
				href: '/status',
				name: 'Runtime & Build Information',
			},
			{
				href: '/flags',
				name: 'Command-Line Flags',
			},
			{
				href: '/config',
				name: 'Configuration',
			},
		],
		name: 'Status',
	},
	{
		children: [
			{
				href: ArgusGitHubRepo,
				name: 'GitHub (source)',
			},
			{
				href: `${ArgusGitHubRepo}/issues`,
				name: 'Report an issue/feature request',
			},
			{
				href: 'https://release-argus.io/docs',
				name: 'Docs',
			},
		],
		name: 'Help',
	},
];

/**
 * The navbar for the app.
 */
const Header = () => {
	const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
	const [expandedItem, setExpandedItem] = useState<string | null>(null);

	const closeMobileMenu = () => {
		setMobileMenuOpen(false);
		setExpandedItem(null);
	};

	return (
		<header className="w-full">
			<div className="flex h-16 w-full shrink-0 items-center justify-between p-4">
				<span className="flex size-full flex-row">
					<Button
						className="mr-2 size-8 lg:hidden"
						onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
						size="icon-md"
						variant="outline"
					>
						<Menu
							className={cn(
								'transition-transform duration-200',
								mobileMenuOpen && 'rotate-90',
							)}
						/>
						<span className="sr-only">Toggle navigation menu</span>
					</Button>

					<Link
						className="mr-2 flex h-full items-center gap-2 rounded-md outline-none focus-visible:outline-1 focus-visible:ring-[3px] focus-visible:ring-ring/50"
						onClick={closeMobileMenu}
						to="/approvals"
					>
						<img alt="" className="h-full w-auto" src="favicon.svg" />
						<h4 className="my-auto scroll-m-20 font-semibold text-xl tracking-tight">
							Argus
						</h4>
					</Link>
					<NavigationMenu delayDuration={0} viewport={false}>
						<NavigationMenuList className="hidden gap-2 lg:flex">
							{headerOptions.map((option) =>
								option.children ? (
									<NavigationMenuItem key={option.name}>
										<NavigationMenuTrigger>{option.name}</NavigationMenuTrigger>
										<NavigationMenuContent className="z-100">
											{option.children.map((child) => {
												const newTab = child.href.startsWith('http');
												return (
													<NavigationMenuLink asChild key={child.name}>
														<Link
															className="whitespace-nowrap"
															rel={newTab ? 'noreferrer noopener' : undefined}
															target={newTab ? '_blank' : undefined}
															to={child.href}
														>
															{child.name}
														</Link>
													</NavigationMenuLink>
												);
											})}
										</NavigationMenuContent>
									</NavigationMenuItem>
								) : (
									<NavigationMenuItem key={option.name}>
										<NavigationMenuLink asChild>
											<Link
												className={cn(
													'group inline-flex h-9 w-max items-center rounded-md px-4 py-2',
													'font-medium text-sm',
												)}
												to={option.href}
											>
												{option.name}
											</Link>
										</NavigationMenuLink>
									</NavigationMenuItem>
								),
							)}
						</NavigationMenuList>
					</NavigationMenu>
				</span>
				<span>
					<ThemeModeToggle />
				</span>
			</div>

			<Collapsible onOpenChange={setMobileMenuOpen} open={mobileMenuOpen}>
				<CollapsibleContent
					className={cn(
						'overflow-hidden lg:hidden',
						'transition-all duration-300 ease-in-out',
						'data-[state=closed]:animate-collapsible-up data-[state=open]:animate-collapsible-down',
					)}
				>
					{headerOptions.map((option) => (
						<div className="last:border-b" key={option.name}>
							{option.children ? (
								<Collapsible
									onOpenChange={() =>
										setExpandedItem((prev) =>
											prev === option.name ? null : option.name,
										)
									}
									open={expandedItem === option.name}
								>
									<CollapsibleTrigger asChild>
										<Button
											className="!px-4 flex w-full items-center justify-between py-2 text-sm"
											variant="ghost"
										>
											{option.name}
											<ChevronDown
												className={cn(
													'transition-transform duration-200',
													expandedItem === option.name && 'rotate-180',
												)}
											/>
										</Button>
									</CollapsibleTrigger>
									<CollapsibleContent
										className={cn(
											'overflow-hidden transition-all duration-200 ease-in-out',
											'data-[state=closed]:animate-collapsible-up',
											'data-[state=open]:animate-collapsible-down',
										)}
									>
										{option.children.map((child) => {
											const newTab = child.href.startsWith('http');
											return (
												<Button
													asChild
													className="w-full justify-start"
													key={child.name}
													variant="ghost"
												>
													<Link
														className="px-8 py-2 text-sm"
														onClick={closeMobileMenu}
														rel={newTab ? 'noreferrer noopener' : undefined}
														target={newTab ? '_blank' : undefined}
														to={child.href}
													>
														{child.name}
													</Link>
												</Button>
											);
										})}
									</CollapsibleContent>
								</Collapsible>
							) : (
								<Button
									asChild
									className="w-full justify-start"
									key={option.name}
									variant="ghost"
								>
									<Link
										className="px-4 py-2 text-sm"
										onClick={closeMobileMenu}
										to={option.href}
									>
										{option.name}
									</Link>
								</Button>
							)}
						</div>
					))}
				</CollapsibleContent>
			</Collapsible>
		</header>
	);
};

export default Header;
