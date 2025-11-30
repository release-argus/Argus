import {
	type DragEndEvent,
	KeyboardSensor,
	PointerSensor,
	useSensor,
	useSensors,
} from '@dnd-kit/core';
import { arrayMove, sortableKeyboardCoordinates } from '@dnd-kit/sortable';
import { useCallback, useState } from 'react';
import { useWebSocket } from '@/contexts/websocket';
import { mapRequest } from '@/utils/api/types/api-request-handler';

/**
 * Manage sortable services.
 *
 * @returns An object containing:
 * - `sensors`: the sensors used for drag and drop functionality.
 * - `handleDragEnd`: function to handle the end of a drag event.
 * - `handleSaveOrder`: function to save the new order of services.
 * - `hasOrderChanged`: boolean indicating if the order has changed.
 * - `resetOrder`: function to reset the order to its original state.
 */
export const useSortableServices = () => {
	const { monitorData, setMonitorData } = useWebSocket();
	// Track the original order of services.
	const [originalOrder, setOriginalOrder] = useState<string[] | null>(null);

	const sensors = useSensors(
		useSensor(PointerSensor),
		useSensor(KeyboardSensor, {
			coordinateGetter: sortableKeyboardCoordinates,
		}),
	);

	// Handle the end of a drag event.
	const handleDragEnd = useCallback(
		(event: DragEndEvent) => {
			const { active, over } = event;
			if (!over || active.id === over.id) return;

			if (!originalOrder || originalOrder.length !== monitorData.order.length) {
				setOriginalOrder(monitorData.order);
			}

			const oldIndex = monitorData.order.indexOf(active.id as string);
			const newIndex = monitorData.order.indexOf(over.id as string);
			if (oldIndex === -1 || newIndex === -1) return;

			const newOrder = arrayMove(monitorData.order, oldIndex, newIndex);
			setMonitorData({
				order: newOrder,
				page: 'APPROVALS',
				sub_type: 'ORDER',
				type: 'SERVICE',
			});
		},
		[monitorData.order, setMonitorData, originalOrder],
	);

	// Send the new ordering to the API.
	// biome-ignore lint/correctness/useExhaustiveDependencies: resetOrder stable with monitorData.order.
	const handleSaveOrder = useCallback(async () => {
		await mapRequest('SERVICE_ORDER_PUT', { order: monitorData.order })
			.then(() => {
				setOriginalOrder(monitorData.order);
			})
			.catch((error: unknown) => {
				console.error('Failed to save order:', error);
				resetOrder();
			});
	}, [monitorData.order]);

	// Order changed if original order not empty, and the new ordering differs.
	const hasOrderChanged =
		originalOrder !== null &&
		originalOrder.length > 0 &&
		!originalOrder.every((id, index) => id === monitorData.order[index]);

	// Reset the order to its original state.
	const resetOrder = useCallback(() => {
		setMonitorData({
			order: originalOrder ?? monitorData.order,
			page: 'APPROVALS',
			sub_type: 'ORDER',
			type: 'SERVICE',
		});
	}, [originalOrder, monitorData.order, setMonitorData]);

	return {
		handleDragEnd,
		handleSaveOrder,
		hasOrderChanged,
		resetOrder,
		sensors,
	};
};
