import { useMutation, useQuery } from '@tanstack/react-query';
import { formatRelative } from 'date-fns';
import { use, useCallback, useEffect, useMemo, useReducer } from 'react';
import { pluralise } from '@/components/generic/util';
import { ModalList } from '@/components/modals/action-release/list';
import { Button } from '@/components/ui/button';
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from '@/components/ui/dialog';
import Tip from '@/components/ui/tip';
import { ModalContext } from '@/contexts/modal';
import { addMessageHandler, removeMessageHandler } from '@/contexts/websocket';
import { QUERY_KEYS } from '@/lib/query-keys';
import reducerActionModal from '@/reducers/action-release';
import type { WebSocketResponse } from '@/types/websocket';
import { dateIsAfterNow, isEmptyObject } from '@/utils';
import { mapRequest } from '@/utils/api/types/api-request-handler';
import type {
	ActionAPIType,
	CommandSummaryListType,
	CommandSummaryType,
	WebHookSummaryListType,
	WebHookSummaryType,
} from '@/utils/api/types/config/summary';
import { isNonEmptyObject } from '@/utils/is-empty';

/**
 * Creates an object with the given number of entries, each with a unique ID and a default next_runnable date.
 *
 * @param count - The number of entries to create.
 */
const createInitialRecord = (count: number) =>
	Object.fromEntries(
		Array.from({ length: count }, (_, i) => [
			' '.repeat(i),
			{ loading: true, next_runnable: '0001-01-01T00:00:00Z' },
		]),
	);

/**
 * @param serviceName - The service name.
 * @param sentCommands - Commands that have sent.
 * @param sentWebHooks - WebHooks that have sent.
 * @returns Whether this service is sending any commands or webhooks.
 */
const isSendingService = (
	serviceName: string,
	sentCommands: string[],
	sentWebHooks: string[],
) => {
	const prefixStr = `${serviceName} `;
	for (const id of sentCommands) {
		if (id.startsWith(prefixStr)) return true;
	}
	for (const id of sentWebHooks) {
		if (id.startsWith(prefixStr)) return true;
	}
	return false;
};

/**
 * @param item - The command/webhook to check.
 * @param allSuccessful - Whether all commands/webhooks have succeeded.
 * @returns Whether we are past the time this command/webhook is next runnable.
 */
const isActionRunnable = (
	item: CommandSummaryType | WebHookSummaryType,
	allSuccessful: boolean,
): boolean => {
	const isScheduledForFuture =
		item.next_runnable && dateIsAfterNow(item.next_runnable);

	const allSucceededOrThisFailed = allSuccessful || item.failed === false;

	// Runnable if not scheduled for the future,
	// and either all have succeeded, or this hasn't.
	return !isScheduledForFuture && allSucceededOrThisFailed;
};

/**
 * @returns The 'Action release' modal, which allows the user to send/retry
 * sending webhooks or commands as long as they are runnable.
 */
const ActionReleaseModal = () => {
	// modal.actionType:
	//   RESEND - 0 WebHooks failed. 'Resend' Modal.
	//   SEND   - Send WebHooks for this new version. 'New release' Modal.
	//   SKIP   - Release not wanted. 'Skip' Modal.
	//   RETRY  - 1+ WebHooks failed sending. 'Retry' Modal.
	const { modal, hideModal: hideModalDialog } = use(ModalContext);
	const [modalData, setModalData] = useReducer(reducerActionModal, {
		commands: {},
		sentC: [],
		sentWH: [],
		service_id: '',
		webhooks: {},
	});

	// biome-ignore lint/correctness/useExhaustiveDependencies: setModal stable.
	const hideModal = useCallback(() => {
		setModalData({ page: 'APPROVALS', sub_type: 'RESET', type: 'ACTION' });
		hideModalDialog();
	}, []);

	// Handle 'deployed version' becoming latest when there is no 'deployed version' check by closing the modal.
	// biome-ignore lint/correctness/useExhaustiveDependencies: hideModal stable.
	useEffect(() => {
		if (
			// Allow 'resend' and 'edit'/'create' modals to stay open.
			!['RESEND', 'EDIT'].includes(modal.actionType) &&
			modal.service.status?.deployed_version &&
			modal.service.status.latest_version &&
			modal.service.status.deployed_version ===
				modal.service.status.latest_version
		) {
			hideModal();
		}
	}, [modal.actionType, modal.service.status]);

	const isSkip = modal.actionType.startsWith('SKIP');
	// biome-ignore lint/correctness/useExhaustiveDependencies: modal.service.command stable with modal.service.id.
	const stats = useMemo(() => {
		const isSending = isSendingService(
			modal.service.id,
			modalData.sentC,
			modalData.sentWH,
		);

		// Whether unspecific actions can be sent
		const hasRunnableCommand = Object.values(modalData.commands).some(
			(command) =>
				!command.next_runnable || !dateIsAfterNow(command.next_runnable),
		);
		const hasRunnableWebhook = Object.values(modalData.webhooks).some(
			(webhook) =>
				!webhook.next_runnable || !dateIsAfterNow(webhook.next_runnable),
		);

		const canSendUnspecific =
			isSkip ||
			(!isSending &&
				isEmptyObject(modalData.commands) &&
				isEmptyObject(modalData.webhooks)) ||
			hasRunnableCommand ||
			hasRunnableWebhook;

		// Action text.
		const commandCount = modal.service.command ?? 0;
		const webhookCount = modal.service.webhook ?? 0;
		const action =
			commandCount && webhookCount
				? 'Commands and WebHooks'
				: commandCount
					? 'Commands'
					: 'WebHooks';

		// Text mappings.
		const textMap: Record<
			string,
			{ title: string; ariaLabel: string; buttonText: string }
		> = {
			DEFAULT: {
				ariaLabel: 'Skip this release',
				buttonText: 'Skip release',
				title: '',
			},
			RESEND: {
				ariaLabel: `Resend the ${action}`,
				buttonText: 'Resend all',
				title: `Resend the ${action}?`,
			},
			RETRY: {
				ariaLabel: `Retry the ${action}`,
				buttonText: 'Retry all failed',
				title: `Retry the ${action}?`,
			},
			SEND: {
				ariaLabel: `Send the ${action} to upgrade`,
				buttonText: 'Confirm',
				title: `Send the ${action} to upgrade?`,
			},
			SKIP: {
				ariaLabel: 'Skip this release',
				buttonText: 'Skip release',
				title: `Skip this release? (don't send any ${action.replace(
					' and',
					' or',
				)})`,
			},
			SKIP_NO_WH: {
				ariaLabel: 'Skip this release',
				buttonText: 'Skip release',
				title: 'Skip this release?',
			},
		};

		const { title, ariaLabel, buttonText } =
			textMap[modal.actionType] || textMap.DEFAULT;

		const confirmButtonText = canSendUnspecific ? buttonText : 'Done';

		return {
			ariaLabel,
			canSendUnspecific,
			confirmButtonText,
			isSending,
			title,
		};
	}, [modal.actionType, modal.service.id, modalData]);

	const { mutate } = useMutation({
		mutationFn: (data: {
			target: string;
			serviceID: string;
			isWebHook: boolean;
			unspecificTarget: boolean;
		}) =>
			mapRequest('ACTION_SEND', {
				serviceID: data.serviceID,
				target: data.target,
			}),
		onMutate: (data) => {
			if (data.target === 'ARGUS_SKIP') return;

			let commandData: CommandSummaryListType = {};
			let webhookData: WebHookSummaryListType = {};
			if (data.unspecificTarget) {
				// All Commands/WebHooks have sent successfully.
				const allSuccessful =
					Object.keys(modalData.commands).every(
						(command_id) => modalData.commands[command_id].failed === false,
					) &&
					Object.keys(modalData.webhooks).every(
						(webhook_id) => modalData.webhooks[webhook_id].failed === false,
					);

				// Send these commands.
				for (const [commandID, command] of Object.entries(modalData.commands)) {
					if (isActionRunnable(command, allSuccessful))
						commandData[commandID] = {};
				}

				// Send these webhooks.
				for (const [webhookID, webhook] of Object.entries(modalData.webhooks)) {
					if (isActionRunnable(webhook, allSuccessful))
						webhookData[webhookID] = {};
				}
				// Targeting specific command/webhook.
			} else if (data.isWebHook) {
				webhookData = { [data.target.slice('webhook_'.length)]: {} };
			} else {
				commandData = { [data.target.slice('command_'.length)]: {} };
			}

			setModalData({
				command_data: commandData,
				page: 'APPROVALS',
				service_data: { id: modal.service.id, loading: false },
				sub_type: 'SENDING',
				type: 'ACTION',
				webhook_data: webhookData,
			});
		},
	});

	// biome-ignore lint/correctness/useExhaustiveDependencies: hideModal stable.
	const onClickAcknowledge = useCallback(
		(target: string, isWebHook?: boolean) => {
			const unspecificTarget = [
				'ARGUS_ALL',
				'ARGUS_FAILED',
				'ARGUS_SKIP',
			].includes(target);

			// Disallow unspecific non-skip targets whilst sending this service.
			if (
				!(
					!stats.canSendUnspecific &&
					unspecificTarget &&
					target !== 'ARGUS_SKIP'
				)
			) {
				let approveTarget = target;
				if (!unspecificTarget)
					if (isWebHook) approveTarget = `webhook_${target}`;
					else approveTarget = `command_${target}`;
				mutate({
					isWebHook: isWebHook === true,
					serviceID: modal.service.id,
					target: approveTarget,
					unspecificTarget: unspecificTarget,
				});
			}

			if (unspecificTarget) hideModal();
		},
		[modal.service, stats.canSendUnspecific],
	);

	// Query for the Commands/WebHooks for the service.
	const { data, isFetching } = useQuery<ActionAPIType>({
		enabled: modal.actionType !== 'EDIT' && modal.service.id !== '',
		placeholderData: {
			command: modal.service.command
				? createInitialRecord(modal.service.command)
				: {},
			webhook: modal.service.webhook
				? createInitialRecord(modal.service.webhook)
				: {},
		},
		queryFn: () => mapRequest('ACTION_GET', { serviceID: modal.service.id }),
		queryKey: QUERY_KEYS.SERVICE.ACTIONS(modal.service.id),
		refetchOnMount: 'always',
		staleTime: 0,
	});

	// Set the modal data from the query data.
	// biome-ignore lint/correctness/useExhaustiveDependencies: modal.service.id stable with data.
	useEffect(() => {
		setModalData({
			command_data: data?.command,
			page: 'APPROVALS',
			service_data: { id: modal.service.id },
			sub_type: 'REFRESH',
			type: 'ACTION',
			webhook_data: data?.webhook,
		});
	}, [data]);

	// Catch WebSocket messages that impact the modal.
	useEffect(() => {
		if (modal.actionType !== 'EDIT' && modal.service.id !== '') {
			// Handler to listen to WebSocket messages.
			const handler = (event: WebSocketResponse) => {
				if (['ACTIONS', 'COMMAND', 'WEBHOOK'].includes(event.type))
					setModalData(event);
			};
			addMessageHandler('action-modal', { handler });
		}

		return () => removeMessageHandler('action-modal');
	}, [modal.actionType, modal.service.id]);

	return (
		<Dialog
			onOpenChange={(open) => {
				if (!open) hideModal();
			}}
			open={!['', 'EDIT'].includes(modal.actionType)}
		>
			<DialogContent className="max-w-lg sm:max-w-2xl">
				<DialogHeader>
					<DialogTitle>
						<strong>{stats.title}</strong>
					</DialogTitle>
					<DialogDescription className="font-semibold">
						<strong>{modal.service.id}</strong>
						{modal.actionType === 'RESEND'
							? ` - ${modal.service.status?.latest_version ?? 'Unknown'}`
							: ''}
					</DialogDescription>
				</DialogHeader>
				{modal.actionType !== 'RESEND' && (
					<div className="flex flex-col gap-0">
						<Tip
							className="w-fit"
							content={
								modal.service.status?.deployed_version_timestamp ? (
									<p>
										{formatRelative(
											new Date(modal.service.status.deployed_version_timestamp),
											new Date(),
										)}
									</p>
								) : (
									<p>Unknown</p>
								)
							}
						>
							<p>
								{`${isSkip ? 'Stay on' : 'From'}: ${
									modal.service.status?.deployed_version ?? 'Unknown'
								}`}
							</p>
						</Tip>
						<Tip
							className="w-fit"
							content={
								modal.service.status?.latest_version_timestamp ? (
									<p>
										{formatRelative(
											new Date(modal.service.status.latest_version_timestamp),
											new Date(),
										)}
									</p>
								) : (
									<p>Unknown</p>
								)
							}
							contentProps={{ side: 'bottom' }}
						>
							{
								<p>
									{`${isSkip ? 'Skip' : 'To'}: ${
										modal.service.status?.latest_version ?? 'Unknown'
									}`}
								</p>
							}
						</Tip>
					</div>
				)}
				{isNonEmptyObject(data?.command) && (
					<div className="flex flex-col gap-1">
						<strong>
							{pluralise('Command', Object.keys(data.command).length)}:
						</strong>
						<ModalList
							data={modalData.commands}
							itemType="COMMAND"
							modalType={modal.actionType}
							onClickAcknowledge={onClickAcknowledge}
							sent={modalData.sentC}
							serviceID={modal.service.id}
						/>
					</div>
				)}
				{isNonEmptyObject(data?.webhook) && (
					<div className="flex flex-col gap-1">
						<strong>
							{pluralise('WebHook', Object.keys(data.webhook).length)}:
						</strong>
						<ModalList
							data={modalData.webhooks}
							itemType="WEBHOOK"
							modalType={modal.actionType}
							onClickAcknowledge={onClickAcknowledge}
							sent={modalData.sentWH}
							serviceID={modal.service.id}
						/>
					</div>
				)}
				<DialogFooter>
					<Button
						hidden={!stats.canSendUnspecific}
						id="modal-cancel"
						onClick={() => hideModal()}
						variant="secondary"
					>
						Cancel
					</Button>
					<Button
						aria-label={stats.ariaLabel}
						disabled={isFetching || (!isSkip && stats.isSending)}
						id="modal-action"
						onClick={() => {
							if (isSkip && !stats.canSendUnspecific) {
								hideModal();
								return;
							}
							switch (modal.actionType) {
								case 'RESEND':
									onClickAcknowledge('ARGUS_ALL');
									break;
								case 'SEND':
								case 'RETRY':
									onClickAcknowledge('ARGUS_FAILED');
									break;
								case 'SKIP':
								case 'SKIP_NO_WH':
									onClickAcknowledge('ARGUS_SKIP');
									break;
							}
						}}
					>
						{stats.confirmButtonText}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
};

export default ActionReleaseModal;
