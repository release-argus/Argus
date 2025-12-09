import EditModeToggle from '@/components/approvals/toolbar/edit-mode-toggle';
import FilterDropdown from '@/components/approvals/toolbar/filter-dropdown';
import SearchBar from '@/components/approvals/toolbar/search-bar';
import TagSelect from '@/components/approvals/toolbar/tag-select';
import ViewToggle from '@/components/approvals/toolbar/view-toggle';
import { ButtonGroup } from '@/components/ui/button-group';
import { TooltipProvider } from '@/components/ui/tooltip';
import { type FC, memo } from 'react';

/**
 * ApprovalsToolbar
 *
 * Toolbar for the 'approvals' view.
 * Manages search, tag filters, edit mode, hide settings, and service order state via URL parameters.
 */

const ApprovalsToolbar: FC = () => (
	<div className="mb-3 flex gap-2 md:gap-3">
		<TooltipProvider>
			<SearchBar />
			<TagSelect />
			<ViewToggle />
			<ButtonGroup>
				<FilterDropdown />
				<EditModeToggle />
			</ButtonGroup>
		</TooltipProvider>
	</div>
);

export default memo(ApprovalsToolbar);
