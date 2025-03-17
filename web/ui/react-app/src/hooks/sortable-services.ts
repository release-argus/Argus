import { Dispatch, useCallback, useState } from 'react';
import {
	KeyboardSensor,
	PointerSensor,
	useSensor,
	useSensors,
} from '@dnd-kit/core';
import { MonitorSummaryType, OrderAPIResponse } from 'types/summary';
import { arrayMove, sortableKeyboardCoordinates } from '@dnd-kit/sortable';

import { WebSocketResponse } from 'types/websocket';
import { fetchJSON } from 'utils';

/**
 * Custom hook to manage sortable services.
 *
 * @param {MonitorSummaryType} monitorData - The current monitor data containing the order of services.
 * @param {Dispatch<WebSocketResponse>} setMonitorData - Reducer function to update the monitor data.
 *
 * @returns {Object} An object containing:
 * - `sensors`: The sensors used for drag and drop functionality.
 * - `handleDragEnd`: Function to handle the end of a drag event.
 * - `handleSaveOrder`: Function to save the new order of services.
 * - `hasOrderChanged`: Boolean indicating if the order has changed.
 * - `resetOrder`: Function to reset the order to its original state.
 */
export const useSortableServices = (
	monitorData: MonitorSummaryType,
	setMonitorData: Dispatch<WebSocketResponse>,
) => {
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
		(event: any) => {
			const { active, over } = event;
			if (active.id !== over.id) {
				const oldIndex = monitorData.order.indexOf(active.id);
				const newIndex = monitorData.order.indexOf(over.id);

				if (!originalOrder || originalOrder.length !== monitorData.order.length)
					setOriginalOrder(monitorData.order);

				const newOrder = arrayMove(monitorData.order, oldIndex, newIndex);
				setMonitorData({
					page: 'APPROVALS',
					type: 'SERVICE',
					sub_type: 'ORDER',
					order: newOrder,
				});
			}
		},
		[monitorData.order, setMonitorData],
	);

	// Send the new ordering to the API.
	const handleSaveOrder = useCallback(async () => {
		console.log(JSON.stringify({ order: monitorData.order }));
		await fetchJSON<OrderAPIResponse>({
			url: 'api/v1/service/order',
			method: 'PUT',
			body: JSON.stringify({ order: monitorData.order }),
		})
			.then(() => {
				setOriginalOrder(monitorData.order);
			})
			.catch((error) => {
				console.error('Failed to save order:', error);
				resetOrder();
			});
	}, [monitorData.order]);

	// Order has changed if the original order is not empty and the new order is different.
	const hasOrderChanged =
		originalOrder !== null &&
		originalOrder.length > 0 &&
		!originalOrder.every((id, index) => id === monitorData.order[index]);

	// Reset the order to its original state.
	const resetOrder = useCallback(() => {
		setMonitorData({
			page: 'APPROVALS',
			type: 'SERVICE',
			sub_type: 'ORDER',
			order: originalOrder ?? monitorData.order,
		});
	}, [originalOrder, monitorData.order, setMonitorData]);

	return {
		sensors,
		handleDragEnd,
		handleSaveOrder,
		hasOrderChanged,
		resetOrder,
	};
};
