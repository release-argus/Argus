import type { QueryClient } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/query-keys.ts';
import type { WebSocketResponse } from '@/types/websocket.ts';
import type {
	ActionAPIType,
	OrderAPIResponse,
	ServiceSummary,
} from '@/utils/api/types/config/summary.ts';

export type ApprovalsQueryCacheUpdaterParams = {
	/* The React Query client. */
	queryClient: QueryClient;
	/* The WebSocket message to process. */
	msg: WebSocketResponse;
};

/**
 * approvalsQueryCacheUpdater updates React Query caches for the Approvals page
 * in response to WebSocket messages.
 *
 * @param queryClient - The React Query client.
 * @param msg - The WebSocket message to process.
 */
export const approvalsQueryCacheUpdater = ({
	queryClient,
	msg,
}: ApprovalsQueryCacheUpdaterParams) => {
	if (msg.page !== 'APPROVALS') return;

	switch (msg.type) {
		// INIT
		// ORDER
		case 'SERVICE': {
			switch (msg.sub_type) {
				case 'INIT': {
					const id = msg?.service_data?.id;
					if (!id) return;

					queryClient.setQueryData(QUERY_KEYS.SERVICE.SUMMARY_ITEM(id), () => ({
						...msg.service_data,
						loading: false,
					}));
					break;
				}
				case 'ORDER': {
					if (msg.order === undefined) return;

					// Set the order cache.
					queryClient.setQueryData(QUERY_KEYS.SERVICE.ORDER(), () => ({
						order: msg.order as string[],
					}));

					// Ensure each item in the order has a cache entry (placeholder if needed).
					for (const id of msg.order) {
						queryClient.setQueryData<ServiceSummary>(
							QUERY_KEYS.SERVICE.SUMMARY_ITEM(id),
							(oldData) => oldData ?? { id, loading: true },
						);
					}
					break;
				}
			}
			break;
		}

		// NEW
		// INIT
		// QUERY
		// UPDATED
		// ACTION
		case 'VERSION': {
			const id = msg.service_data?.id;
			if (!id) break;

			queryClient.setQueryData<ServiceSummary>(
				QUERY_KEYS.SERVICE.SUMMARY_ITEM(id),
				(_oldData) => {
					const oldData = _oldData ?? { id: id };
					const mergedStatus = {
						...oldData?.status,
						...msg.service_data?.status,
					};

					// Default the approved_version/deployed_version to latest_version.
					if (msg.sub_type === 'INIT') {
						mergedStatus.deployed_version ??=
							msg.service_data?.status?.latest_version;
						mergedStatus.deployed_version_timestamp ??=
							msg.service_data?.status?.latest_version_timestamp;
					}

					return {
						...oldData,
						...msg.service_data,
						loading: false,
						status: mergedStatus,
					};
				},
			);
			break;
		}

		case 'EDIT': {
			const newServiceData = msg.service_data;
			if (!newServiceData) break;

			// old ID when editing existing service.
			const oldID = msg.sub_type;
			const newID = newServiceData.id ?? oldID;
			if (!newID) break;

			const orderData = queryClient.getQueryData<OrderAPIResponse>(
				QUERY_KEYS.SERVICE.ORDER(),
			);
			const newOrder = orderData?.order ?? [];
			if (oldID && oldID !== newID) {
				// Clear caches if another service already has this ID.
				if (newOrder.includes(newID)) {
					queryClient
						.invalidateQueries({
							exact: true,
							queryKey: QUERY_KEYS.SERVICE.ORDER(),
						})
						.catch((err) => {
							console.error('Failed to invalidate queries', err);
						});
					queryClient
						.invalidateQueries({
							queryKey: QUERY_KEYS.SERVICE.SUMMARY_ITEM_BASE,
						})
						.catch((err) => {
							console.error('Failed to invalidate queries', err);
						});
					return;
				}

				// Replace ID in the ORDER array.
				const idx = newOrder.indexOf(oldID);
				if (idx !== -1) newOrder[idx] = newID;
				if (newOrder.length > 0) {
					queryClient.setQueryData(QUERY_KEYS.SERVICE.ORDER(), () => ({
						order: newOrder,
					}));
				}

				// Remove old summary cache.
				queryClient.removeQueries({
					exact: true,
					queryKey: QUERY_KEYS.SERVICE.SUMMARY_ITEM(oldID),
				});
			}

			// Upsert summary for newID.
			queryClient.setQueryData<ServiceSummary>(
				QUERY_KEYS.SERVICE.SUMMARY_ITEM(newID),
				(oldData) => ({
					...oldData,
					...newServiceData,
					loading: false,
					status: {
						...oldData?.status,
						...newServiceData.status,
					},
				}),
			);

			// Ensure ID in order.
			if (!newOrder.includes(newServiceData.id)) {
				queryClient.setQueryData<OrderAPIResponse>(
					QUERY_KEYS.SERVICE.ORDER(),
					(oldData) => ({
						order: [...(oldData?.order ?? []), newID],
					}),
				);
			}

			// Invalidate related caches.
			queryClient
				.invalidateQueries({
					queryKey: QUERY_KEYS.SERVICE.ACTIONS(newID),
				})
				.catch((err) => {
					console.error('Failed to invalidate queries', err);
				});
			queryClient
				.invalidateQueries({
					queryKey: QUERY_KEYS.SERVICE.EDIT_ITEM(newID),
				})
				.catch((err) => {
					console.error('Failed to invalidate queries', err);
				});
			break;
		}

		case 'DELETE': {
			if (!msg.sub_type) break;
			const id = msg.sub_type;

			// Remove from ORDER.
			queryClient.setQueryData<OrderAPIResponse>(
				QUERY_KEYS.SERVICE.ORDER(),
				(oldData) => ({
					order: (oldData?.order ?? []).filter((x) => x !== id),
				}),
			);

			// Remove caches for this service.
			queryClient.removeQueries({
				queryKey: QUERY_KEYS.SERVICE.BASE(id),
			});
			break;
		}

		// EVENT
		case 'COMMAND':
		case 'WEBHOOK': {
			const queryKey = QUERY_KEYS.SERVICE.ACTIONS(msg.service_data?.id);

			// {
			//  "page":"APPROVALS",
			//  "type":"WEBHOOK",
			//  "sub_type":"EVENT",
			//  "service_data":{"id":"release-argus/Argus"},
			//  "webhook_data":{"awx":{"failed":false,"next_runnable":"2025-12-01T01:02:03.12345678Z"}}
			// }
			queryClient.setQueryData<ActionAPIType>(queryKey, (_prevData) => {
				const prevData = _prevData ?? { command: {}, webhook: {} };

				const mergedCommands = prevData.command;
				if (msg.command_data) {
					for (const [id, data] of Object.entries(msg.command_data)) {
						mergedCommands[id] = {
							...mergedCommands[id],
							failed: data.failed,
							next_runnable: data.next_runnable,
						};
					}
				}

				const mergedWebhooks = prevData.webhook;
				if (msg.webhook_data) {
					for (const [id, data] of Object.entries(msg.webhook_data)) {
						mergedWebhooks[id] = {
							...mergedWebhooks[id],
							failed: data.failed,
							next_runnable: data.next_runnable,
						};
					}
				}

				return {
					...prevData,
					command: mergedCommands,
					webhook: mergedWebhooks,
				};
			});
			break;
		}

		default:
			break;
	}
};

export default approvalsQueryCacheUpdater;
