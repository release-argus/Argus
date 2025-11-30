import type { RefObject } from 'react';
import { toast } from 'sonner';
import type { WebSocketResponse } from '@/types/websocket';
import type {
	MonitorSummaryType,
	ServiceSummary,
} from '@/utils/api/types/config/summary';

/**
 * Adds a notification based on the event type and subtype
 *
 * @param event - The WebSocket message that may trigger a notification.
 * @param params - Extra parameters for handling notifications.
 * @param params.monitorData - The Service data ref from the WebSocket.
 */
export const handleNotifications = (
	event: WebSocketResponse,
	params?: { monitorData: RefObject<MonitorSummaryType> },
) => {
	if (event.page !== 'APPROVALS' || !params) return;
	// APPROVALS

	const serviceData = (event as { service_data?: ServiceSummary }).service_data;
	if (serviceData === undefined) return;

	const monitorServiceData = params.monitorData.current.service[
		serviceData.id
	] as ServiceSummary | undefined;
	const serviceName = monitorServiceData?.name ?? serviceData.id;

	// VERSION
	// COMMAND
	// WEBHOOK
	switch (event.type) {
		case 'VERSION':
			// QUERY
			// NEW
			// UPDATED
			// INIT
			// ACTION
			switch (event.sub_type) {
				case 'QUERY':
					break;
				case 'NEW':
					toast.info(serviceName, {
						description: `New version: ${serviceData.status?.latest_version ?? 'Unknown'}`,
						duration: 15000, // 15 seconds.
					});
					break;
				case 'UPDATED':
					toast.success(serviceName, {
						description: `Updated to version '${
							serviceData.status?.deployed_version ?? 'Unknown'
						}'`,
						duration: 15000, // 15 seconds.
					});
					break;
				case 'INIT':
					toast.info(serviceName, {
						description: `Service initialised with version '${
							serviceData.status?.deployed_version ?? 'Unknown'
						}'`,
						duration: 5000, // 5 seconds.
					});
					break;
				case 'ACTION':
					// Notify on 'SKIP'.
					if (serviceData.status?.approved_version) {
						if (serviceData.status.approved_version.startsWith('SKIP_'))
							toast.info(serviceName, {
								description: `Skipped version: ${serviceData.status.approved_version.slice(
									'SKIP_'.length,
								)}`,
								duration: 15000, // 15 seconds.
							});
					}
					break;
				default:
					break;
			}
			break;

		case 'COMMAND':
			if (event.sub_type !== 'EVENT') return;
			// EVENT
			for (const [key, cmd] of Object.entries(event.command_data ?? [])) {
				if (cmd.failed === false) {
					toast.success(serviceName, {
						description: `'${key}' Command ran successfully`,
						duration: 5000,
					});
				} else {
					toast.error(serviceName, {
						description: `'${key}' Command failed`,
						duration: 5000,
					});
				}
			}
			break;
		case 'WEBHOOK':
			if (event.sub_type !== 'EVENT') return;
			// EVENT
			for (const [key, wh] of Object.entries(event.webhook_data ?? [])) {
				if (wh.failed === false) {
					toast.success(serviceName, {
						description: `'${key}' WebHook sent successfully`,
						duration: 5000,
					});
				} else {
					toast.error(serviceName, {
						description: `'${key}' WebHook failed to send`,
						duration: 5000,
					});
				}
			}
			break;
	}
};
