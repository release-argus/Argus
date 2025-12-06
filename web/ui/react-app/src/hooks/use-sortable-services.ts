import {
	type DragEndEvent,
	KeyboardSensor,
	PointerSensor,
	useSensor,
	useSensors,
} from '@dnd-kit/core';
import { arrayMove, sortableKeyboardCoordinates } from '@dnd-kit/sortable';
import { useQueryClient } from '@tanstack/react-query';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useServiceOrder } from '@/hooks/use-service-order.ts';
import { QUERY_KEYS } from '@/lib/query-keys.ts';
import { mapRequest } from '@/utils/api/types/api-request-handler';
import diffLists from '@/utils/diff-lists.ts';

/**
 * Manage sortable services.
 *
 * @returns An object containing:
 * - `sensors`: the sensors used for drag and drop functionality.
 * - `handleDragEnd`: function to handle the end of a drag event.
 * - `handleSaveOrder`: function to save the new order of services.
 * - `hasOrderChanged`: boolean indicating if the order has changed.
 * - `order`: the current order of services.
 * - `resetOrder`: function to reset the order to its original state.
 */
export const useSortableServices = () => {
	const queryClient = useQueryClient();
	const { data: orderData } = useServiceOrder();
	const serverOrder = orderData?.order ?? [];

	// Track the original order of services.
	const [order, setOrder] = useState<string[]>([]);

	// biome-ignore lint/correctness/useExhaustiveDependencies: orderData covers serverOrder.
	useEffect(() => {
		// Initialise from server order once available.
		if (serverOrder.length > 0) {
			// If a service was added/removed, reset the order to the server order.
			const diffServiceIDs = diffLists({ listA: order, listB: serverOrder });
			const applyOrder = !hasOrderChanged || diffServiceIDs;

			if (order.length === 0 || applyOrder) setOrder(serverOrder);
		}
	}, [serverOrder]);

	const sensors = useSensors(
		useSensor(PointerSensor),
		useSensor(KeyboardSensor, {
			coordinateGetter: sortableKeyboardCoordinates,
		}),
	);

	// Handle the end of a drag event.
	// biome-ignore lint/correctness/useExhaustiveDependencies: queryClient stable.
	const handleDragEnd = useCallback(
		(event: DragEndEvent) => {
			const { active, over } = event;
			if (!over || active.id === over.id || order.length === 0) return;

			const oldIndex = order.indexOf(active.id as string);
			const newIndex = order.indexOf(over.id as string);
			if (oldIndex === -1 || newIndex === -1) return;

			const newOrder = arrayMove(order, oldIndex, newIndex);

			// Update local state.
			setOrder(newOrder);
		},
		[order, serverOrder],
	);

	// Reset the order to its original state.
	// biome-ignore lint/correctness/useExhaustiveDependencies: queryClient stable.
	const resetOrder = useCallback(() => {
		const resetTo = serverOrder;
		if (!resetTo || resetTo.length === 0) return;

		setOrder(resetTo);
	}, [serverOrder]);

	// Send the new ordering to the API.
	// biome-ignore lint/correctness/useExhaustiveDependencies: resetOrder stable with order.
	const handleSaveOrder = useCallback(async () => {
		if (!order || order.length === 0) return;
		await mapRequest('SERVICE_ORDER_PUT', { order })
			.then(() => {
				// Sync local state and query cache.
				queryClient.setQueryData(QUERY_KEYS.SERVICE.ORDER(), () => ({
					order: order,
				}));
			})
			.catch((error: unknown) => {
				console.error('Failed to save order:', error);
				resetOrder();
			});
	}, [order, resetOrder]);

	// Order changed if original order not empty, and the new ordering differs.
	const hasOrderChanged = useMemo(
		() =>
			serverOrder.length > 0 &&
			order.length > 0 &&
			serverOrder.some((id, index) => id !== order[index]),
		[serverOrder, order],
	);

	return {
		handleDragEnd,
		handleSaveOrder,
		hasOrderChanged,
		order,
		resetOrder,
		sensors,
	};
};
