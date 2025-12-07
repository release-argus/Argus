import { Button } from '@/components/ui/button.tsx';
import { DropdownMenuCheckboxItem } from '@/components/ui/dropdown-menu';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { GripVertical } from 'lucide-react';
import type { FC } from 'react';

type DropdownMenuCheckboxItemSortableProps = {
	/* ID of the checkbox */
	id: string;
	/* Label for the checkbox */
	label: string;
	/* Checked state of the checkbox */
	checked: boolean;
	/* Function to call on checked state change */
	onCheckedChange: (checked: boolean) => void;
};

/**
 * A checkbox item that can be dragged around in a dropdown menu.
 *
 * @param id - ID of the checkbox
 * @param label - Label for the checkbox
 * @param checked - Checked state of the checkbox
 * @param onCheckedChange - Function to call on checked state change
 */
export const DropdownMenuCheckboxItemSortable: FC<
	DropdownMenuCheckboxItemSortableProps
> = ({ id, label, checked, onCheckedChange }) => {
	const {
		attributes,
		listeners,
		setNodeRef,
		transform,
		transition,
		isDragging,
	} = useSortable({ id: id });

	const style = {
		opacity: isDragging ? 0.4 : 1,
		transform: CSS.Transform.toString(transform),
		transition,
	};

	return (
		<div className="flex items-center" ref={setNodeRef} style={style}>
			<DropdownMenuCheckboxItem
				checked={checked}
				className="flex-1"
				onCheckedChange={onCheckedChange}
			>
				{label}
			</DropdownMenuCheckboxItem>
			<Button
				{...listeners}
				{...attributes}
				aria-label="Drag handle"
				className="cursor-grab touch-none px-0! py-2 text-muted-foreground"
				size="sm"
				variant="ghost"
			>
				<GripVertical className="h-4 w-4"/>
			</Button>
		</div>
	);
};
