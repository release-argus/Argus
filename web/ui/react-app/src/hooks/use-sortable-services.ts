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
import { toast } from 'sonner';
import { useServiceOrder } from '@/hooks/use-service-order';
import { QUERY_KEYS } from '@/lib/query-keys';
import { mapRequest } from '@/utils/api/types/api-request-handler';
import diffLists from '@/utils/diff-lists';

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
	const { data: orderData, isSuccess: haveOrderData } = useServiceOrder();
	const serverOrder = orderData?.order ?? [];

	// Track the original order of services.
	const [order, setOrder] = useState<string[]>([]);

	// biome-ignore lint/correctness/useExhaustiveDependencies: orderData covers serverOrder.
	useEffect(() => {
		// Initialise from server order once available.
		if (haveOrderData) {
			// If a service was added/removed, reset the order to the server order.
			const diffServiceIDs = diffLists({ listA: order, listB: serverOrder });
			const shouldApplyOrder = !hasOrderChanged || diffServiceIDs;

			if (order.length === 0 || shouldApplyOrder) setOrder(serverOrder);
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
	const resetOrder = useCallback(() => {
		const resetTo = serverOrder;
		if (!resetTo || resetTo.length === 0) return;

		setOrder(resetTo);
	}, [serverOrder]);

	// Apply a full new order (expects the complete list of service IDs).
	// Ignore lists of differing length to the current server order.
	// biome-ignore lint/correctness/useExhaustiveDependencies: haveOrderData stable with serverOrder.
	const applyOrder = useCallback(
		(ids: string[]) => {
			if (!ids || ids.length === 0 || !haveOrderData) return;

			// Validate it is a full permutation of serverOrder.
			if (ids.length !== serverOrder.length) return;
			const a = [...ids].toSorted((a, b) => a.localeCompare(b));
			const b = [...serverOrder].toSorted((a, b) => a.localeCompare(b));
			for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return;

			setOrder(ids);
		},
		[serverOrder],
	);

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
				toast.error('Failed to save order.', {
					description: `Error: ${error instanceof Error ? error.message : String(error)}`,
				});
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
		applyOrder,
		handleDragEnd,
		handleSaveOrder,
		hasOrderChanged,
		order,
		resetOrder,
		sensors,
	};
};
