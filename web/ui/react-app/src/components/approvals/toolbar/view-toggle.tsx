import type { FC } from 'react';
import { useToolbar } from '@/components/approvals/toolbar/toolbar-context';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import {
	approvalsToolbarViewOptions,
	type ToolbarViewOption,
} from '@/constants/toolbar';

/**
 * ViewToggle
 *
 * A toggle component for toggling view mode on services.
 */
const ViewToggle: FC = () => {
	const { values, setView } = useToolbar();
	const currentValue = values.view;

	return (
		<ToggleGroup
			className="hidden sm:flex"
			onValueChange={(val: ToolbarViewOption | '') => {
				if (!val) return;
				setView(val);
			}}
			type="single"
			value={currentValue}
			variant="outline"
		>
			{approvalsToolbarViewOptions.map(({ icon: Icon, value, label }) => (
				<ToggleGroupItem
					aria-label={`${label} layout`}
					key={label}
					value={value}
				>
					<Icon className="h-4 w-4" />
				</ToggleGroupItem>
			))}
		</ToggleGroup>
	);
};

export default ViewToggle;
