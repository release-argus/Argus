import { SearchIcon, X } from 'lucide-react';
import { type FC, useEffect, useRef } from 'react';
import { useToolbar } from '@/components/approvals/toolbar/toolbar-context';
import { Button } from '@/components/ui/button';
import {
	InputGroup,
	InputGroupAddon,
	InputGroupButton,
	InputGroupInput,
} from '@/components/ui/input-group';
import { Kbd } from '@/components/ui/kbd';
import { Separator } from '@/components/ui/separator';
import { useIsMobile } from '@/hooks/use-mobile';

/**
 * Toolbar input for filtering services by name.
 *
 * Supports '/' to focus and 'Escape' to clear.
 */
const SearchBar: FC = () => {
	const { values, setSearch } = useToolbar();
	const search = values.search ?? '';
	const searchInputRef = useRef<HTMLInputElement | null>(null);
	const isMobile = useIsMobile();

	// Event listeners for focus/clear.
	useEffect(() => {
		const handleKeyPress = (event: KeyboardEvent) => {
			if (
				event.target instanceof HTMLInputElement ||
				event.target instanceof HTMLTextAreaElement
			)
				return;

			if (event.key === '/') {
				event.preventDefault();
				searchInputRef.current?.focus();
			}
		};

		const handleEscape = (event: KeyboardEvent) => {
			if (
				event.key === 'Escape' &&
				searchInputRef.current &&
				document.activeElement === searchInputRef.current
			) {
				searchInputRef.current.blur();
				setSearch('');
			}
		};

		globalThis.addEventListener('keydown', handleKeyPress);
		globalThis.addEventListener('keydown', handleEscape);

		return () => {
			globalThis.removeEventListener('keydown', handleKeyPress);
			globalThis.removeEventListener('keydown', handleEscape);
		};
	}, [setSearch]);

	return (
		<InputGroup>
			<InputGroupInput
				aria-label="Search and filter services by name"
				onChange={(e) => {
					setSearch(e.target.value);
				}}
				placeholder="Filter services"
				ref={searchInputRef}
				type="text"
				value={search}
			/>
			<InputGroupAddon>
				<SearchIcon />
			</InputGroupAddon>
			{!isMobile && (
				<InputGroupAddon align="inline-end">
					<Kbd>/</Kbd>
				</InputGroupAddon>
			)}
			{search && (
				<InputGroupAddon align="inline-end" className="!py-0 !pr-1.5 gap-0">
					<InputGroupButton
						aria-label="Clear search"
						className="order-last rounded-l-none border-l text-muted-foreground hover:text-foreground focus-visible:text-foreground"
						onClick={() => setSearch('')}
						size="icon-md"
						variant="ghost"
					>
						<X />
					</InputGroupButton>
				</InputGroupAddon>
			)}
		</InputGroup>
	);
};

export default SearchBar;
