import { Service } from '@/components/approvals';
import type { ServiceSummary } from '@/utils/api/types/config/summary';
import { closestCenter, DndContext, type DragEndEvent, type SensorOptions, } from '@dnd-kit/core';
import type { SensorDescriptor } from '@dnd-kit/core/dist/sensors/types';
import { SortableContext } from '@dnd-kit/sortable';
import type { FC } from 'react';

type GridLayoutProps = {
	/* The list of service summaries for the grid. */
	services: ServiceSummary[];
	/* A flag indicating whether the layout is in edit mode. */
	editMode: boolean;
	/* The current order of services. */
	order: string[];
	/* An array of sensor descriptors for drag-and-drop operations. */
	sensors: SensorDescriptor<SensorOptions>[];
	/* A function invoked after the drag-and-drop interaction on services is completed. */
	handleDragEnd: (event: DragEndEvent) => void;
};

/**
 * A functional component that renders a drag-sortable grid layout for displaying services.
 *
 * @param services - The list of service summaries for the grid.
 * @param editMode - A flag indicating whether the layout is in edit mode.
 * @param order - The current order of services.
 * @param sensors - An array of sensor descriptors for drag-and-drop operations.
 * @param handleDragEnd - A function invoked after a drag-and-drop interaction is completed.
 * @returns The rendered grid layout containing the given services in the order specified with drag-and-drop capabilities.
 */
export const GridLayout: FC<GridLayoutProps> = ({
	services,
	editMode,
	order,
	sensors,
	handleDragEnd,
}) => {
	return (
		<div className="grid gap-4 [grid-template-columns:repeat(auto-fill,minmax(17.5rem,1fr))]">
			<DndContext
				autoScroll={{
					acceleration: 100,
					enabled: true,
					interval: 5,
					threshold: {
						x: 0.2, // Start scrolling when within 20% of the edge.
						y: 0.2,
					},
				}}
				collisionDetection={closestCenter}
				onDragEnd={handleDragEnd}
				sensors={sensors}
			>
				<SortableContext items={order}>
					{services.map((s) => (
						<Service editable={editMode} id={s.id} key={s.id} />
					))}
				</SortableContext>
			</DndContext>
		</div>
	);
};
