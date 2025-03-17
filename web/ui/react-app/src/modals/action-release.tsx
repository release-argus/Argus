import {
	ActionAPIType,
	CommandSummaryListType,
	WebHookSummaryListType,
} from 'types/summary';
import {
	Button,
	Container,
	Modal,
	OverlayTrigger,
	Tooltip,
} from 'react-bootstrap';
import { dateIsAfterNow, fetchJSON, isEmptyObject } from 'utils';
import { useCallback, useContext, useEffect, useMemo, useReducer } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';

import { ModalContext } from 'contexts/modal';
import { ModalList } from 'components/modals/action-release/list';
import { WebSocketResponse } from 'types/websocket';
import { addMessageHandler } from 'contexts/websocket';
import { formatRelative } from 'date-fns';
import { pluralise } from 'components/generic/util';
import reducerActionModal from 'reducers/action-release';
import { useDelayedRender } from 'hooks/delayed-render';

/**
 * Whether the service is sending any commands or webhooks.
 *
 * @param serviceName - The service name.
 * @param sentCommands - The sent commands.
 * @param sentWebHooks - The sent webhooks.
 * @returns Whether the service is sending any commands or webhooks.
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
 * @returns The action release modal, which allows the user to send/retry
 * sending webhooks or commands as long as they are runnable.
 */
const ActionReleaseModal = () => {
	// modal.actionType:
	// RESEND - 0 WebHooks failed. Resend Modal.
	// SEND   - Send WebHooks for this new version. New release Modal.
	// SKIP   - Release not wanted. Skip release Modal.
	// RETRY  - 1+ WebHooks failed to send. Retry send Modal.
	const { handleModal, modal } = useContext(ModalContext);
	const delayedRender = useDelayedRender(250);
	const [modalData, setModalData] = useReducer(reducerActionModal, {
		service_id: '',
		sentC: [],
		sentWH: [],
		webhooks: {},
		commands: {},
	});

	const hideModal = useCallback(() => {
		setModalData({ page: 'APPROVALS', type: 'ACTION', sub_type: 'RESET' });
		handleModal('', { id: '', loading: true });
	}, []);

	// Handle deployed version becoming latest when there's no deployed version check.
	// (close the modal)
	useEffect(() => {
		if (
			// Allow resend and edit/create modals to stay open.
			!['RESEND', 'EDIT'].includes(modal.actionType) &&
			modal.service?.status?.deployed_version &&
			modal.service?.status?.latest_version &&
			modal.service?.status?.deployed_version ===
				modal.service?.status?.latest_version
		)
			hideModal();
	}, [modal.actionType, modal.service?.status]);

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
			(!isSending &&
				isEmptyObject(modalData.commands) &&
				isEmptyObject(modalData.webhooks)) ||
			hasRunnableCommand ||
			hasRunnableWebhook;

		// Action text.
		const commandCount = modal.service?.command ?? 0;
		const webhookCount = modal.service?.webhook ?? 0;
		const action =
			commandCount && webhookCount
				? `${pluralise('Command', commandCount)} and ${pluralise(
						'WebHook',
						webhookCount,
				  )}`
				: commandCount
				? pluralise('Command', commandCount)
				: pluralise('WebHook', webhookCount);

		// Text mappings.
		const textMap: Record<
			string,
			{ title: string; ariaLabel: string; buttonText: string }
		> = {
			RESEND: {
				title: `Resend the ${action}?`,
				ariaLabel: `Resend the ${action}`,
				buttonText: 'Resend all',
			},
			SEND: {
				title: `Send the ${action} to upgrade?`,
				ariaLabel: `Send the ${action} to upgrade`,
				buttonText: 'Confirm',
			},
			SKIP: {
				title: `Skip this release? (don't send any ${action.replace(
					'and',
					'or',
				)})`,
				ariaLabel: 'Skip this release',
				buttonText: 'Skip release',
			},
			SKIP_NO_WH: {
				title: 'Skip this release?',
				ariaLabel: 'Skip this release',
				buttonText: 'Skip release',
			},
			RETRY: {
				title: `Retry the ${action}?`,
				ariaLabel: `Retry the ${action}`,
				buttonText: 'Retry all failed',
			},
			DEFAULT: {
				title: '',
				ariaLabel: 'Skip this release',
				buttonText: 'Skip release',
			},
		};

		const { title, ariaLabel, buttonText } =
			textMap[modal.actionType] || textMap.DEFAULT;

		const confirmButtonText = canSendUnspecific ? buttonText : 'Done';

		return {
			isSending,
			canSendUnspecific,
			title,
			ariaLabel,
			confirmButtonText,
		};
	}, [modal.actionType, modal.service.id, modalData]);

	const { mutate } = useMutation({
		mutationFn: (data: {
			target: string;
			service: string;
			isWebHook: boolean;
			unspecificTarget: boolean;
		}) =>
			fetchJSON({
				url: `api/v1/service/actions/${encodeURIComponent(data.service)}`,
				method: 'POST',
				body: JSON.stringify({ target: data.target }),
			}),
		onMutate: (data) => {
			if (data.target === 'ARGUS_SKIP') return;

			let command_data: CommandSummaryListType | undefined = {};
			let webhook_data: WebHookSummaryListType | undefined = {};
			if (!data.unspecificTarget) {
				// Targeting specific command/webhook.
				if (data.isWebHook)
					webhook_data = { [data.target.slice('webhook_'.length)]: {} };
				else command_data = { [data.target.slice('command_'.length)]: {} };
			} else {
				// All Commands/WebHooks have been sent successfully.
				const allSuccessful =
					Object.keys(modalData.commands).every(
						(command_id) => modalData.commands[command_id].failed === false,
					) &&
					Object.keys(modalData.webhooks).every(
						(webhook_id) => modalData.webhooks[webhook_id].failed === false,
					);

				// sending these command(s).
				for (const command_id in modalData.commands) {
					// skip commands that aren't after next_runnable
					// and commands that have already succeeded if some commands haven't.
					if (
						(modalData.commands[command_id].next_runnable !== undefined &&
							dateIsAfterNow(modalData.commands[command_id].next_runnable)) ||
						(!allSuccessful && modalData.commands[command_id].failed === false)
					)
						continue;
					command_data[command_id] = {};
				}

				// sending these webhook(s).
				for (const webhook_id in modalData.webhooks) {
					// skip webhooks that aren't after next_runnable
					// and webhooks that have already succeeded if some webhooks haven't.
					if (
						(modalData.webhooks[webhook_id].next_runnable !== undefined &&
							dateIsAfterNow(modalData.webhooks[webhook_id].next_runnable)) ||
						(!allSuccessful && modalData.webhooks[webhook_id].failed === false)
					)
						continue;
					webhook_data[webhook_id] = {};
				}
			}

			setModalData({
				page: 'APPROVALS',
				type: 'ACTION',
				sub_type: 'SENDING',
				service_data: { id: modal.service.id, loading: false },
				command_data: command_data,
				webhook_data: webhook_data,
			});
		},
	});

	const onClickAcknowledge = useCallback(
		(target: string, isWebHook?: boolean) => {
			const unspecificTarget = [
				'ARGUS_ALL',
				'ARGUS_FAILED',
				'ARGUS_SKIP',
			].includes(target);

			// Do not allow unspecific non-skip targets if currently sending this service.
			if (
				!(
					!stats.canSendUnspecific &&
					unspecificTarget &&
					target !== 'ARGUS_SKIP'
				)
			) {
				console.log(`Approving ${modal.service.id} - ${target}`);
				let approveTarget = target;
				if (!unspecificTarget)
					if (isWebHook) approveTarget = `webhook_${target}`;
					else approveTarget = `command_${target}`;
				mutate({
					service: modal.service.id,
					target: approveTarget,
					isWebHook: isWebHook === true,
					unspecificTarget: unspecificTarget,
				});
			}

			if (unspecificTarget) hideModal();
		},
		[modal.service, stats.canSendUnspecific],
	);

	// Query for the Commands/WebHooks for the service.
	const { data } = useQuery<ActionAPIType>({
		queryKey: ['actions', { service: modal.service.id }],
		queryFn: () =>
			fetchJSON({
				url: `api/v1/service/actions/${encodeURIComponent(modal.service.id)}`,
			}),
		enabled: modal.actionType !== 'EDIT' && modal.service.id !== '',
		refetchOnMount: 'always',
	});

	// Set the modal data from the query data.
	useEffect(
		() =>
			setModalData({
				page: 'APPROVALS',
				type: 'ACTION',
				sub_type: 'REFRESH',
				service_data: { id: modal.service.id },
				webhook_data: data?.webhook,
				command_data: data?.command,
			}),
		[data],
	);

	// Catch WebSocket messages that impact the modal.
	useEffect(() => {
		if (modal.actionType !== 'EDIT' && modal.service.id !== '') {
			// Handler to listen to WebSocket messages.
			const handler = (event: WebSocketResponse) => {
				if (event && ['ACTIONS', 'COMMAND', 'WEBHOOK'].includes(event.type))
					setModalData(event);
			};
			addMessageHandler('action-modal', handler);
		}
	}, [modal.actionType, modal.service.id]);

	return (
		<Modal
			show={!['', 'EDIT'].includes(modal.actionType)}
			onHide={() => hideModal()}
		>
			<Modal.Header closeButton>
				<Modal.Title>
					<strong>{stats.title}</strong>
				</Modal.Title>
			</Modal.Header>
			<Modal.Body>
				<Container
					fluid
					className="font-weight-bold"
					style={{ paddingLeft: '0px' }}
				>
					<strong>{modal.service.id}</strong>
					{modal.actionType === 'RESEND'
						? ` - ${modal.service?.status?.latest_version ?? 'Unknown'}`
						: ''}
				</Container>
				<>
					{modal.actionType !== 'RESEND' && (
						<>
							<OverlayTrigger
								key="from-version"
								placement="top"
								delay={{ show: 500, hide: 500 }}
								overlay={
									<Tooltip id="tooltip-deployed-version">
										{modal.service?.status?.deployed_version_timestamp ? (
											<>
												{formatRelative(
													new Date(
														modal.service?.status?.deployed_version_timestamp,
													),
													new Date(),
												)}
											</>
										) : (
											<>Unknown</>
										)}
									</Tooltip>
								}
							>
								<p style={{ margin: 0, maxWidth: 'fit-content' }}>
									{`${modal.actionType === 'SKIP' ? 'Stay on' : 'From'}: ${
										modal.service?.status?.deployed_version
									}`}
								</p>
							</OverlayTrigger>
							<OverlayTrigger
								key="to-version"
								placement="bottom"
								delay={{ show: 500, hide: 500 }}
								overlay={
									<Tooltip id="tooltip-latest-version">
										{modal.service?.status?.latest_version_timestamp ? (
											<>
												{formatRelative(
													new Date(
														modal.service?.status?.latest_version_timestamp,
													),
													new Date(),
												)}
											</>
										) : (
											<>Unknown</>
										)}
									</Tooltip>
								}
							>
								<p style={{ margin: 0, maxWidth: 'fit-content' }}>
									{`${modal.actionType === 'SKIP' ? 'Skip' : 'To'}: ${
										modal.service?.status?.latest_version
									}`}
								</p>
							</OverlayTrigger>
						</>
					)}
					{!isEmptyObject(data?.webhook) && (
						<>
							<br />
							<strong>WebHook(s):</strong>
							<ModalList
								itemType="WEBHOOK"
								modalType={modal.actionType}
								serviceID={modal.service.id}
								data={modalData.webhooks}
								sent={modalData.sentWH}
								onClickAcknowledge={onClickAcknowledge}
								delayedRender={delayedRender}
							/>
						</>
					)}
					{!isEmptyObject(data?.command) && (
						<>
							<br />
							<strong>Command(s):</strong>
							<ModalList
								itemType="COMMAND"
								modalType={modal.actionType}
								serviceID={modal.service.id}
								data={modalData.commands}
								sent={modalData.sentC}
								onClickAcknowledge={onClickAcknowledge}
								delayedRender={delayedRender}
							/>
						</>
					)}
				</>
			</Modal.Body>
			<Modal.Footer>
				<Button
					id="modal-cancel"
					variant="secondary"
					hidden={!stats.canSendUnspecific}
					onClick={() => hideModal()}
				>
					Cancel
				</Button>
				<Button
					aria-label={stats.ariaLabel}
					id="modal-action"
					variant="primary"
					onClick={() => {
						if (
							!['SKIP', 'SKIP_NO_WH'].includes(modal.actionType) &&
							!stats.canSendUnspecific
						) {
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
					disabled={modal.actionType !== 'SKIP' && stats.isSending}
				>
					{stats.confirmButtonText}
				</Button>
			</Modal.Footer>
		</Modal>
	);
};

export default ActionReleaseModal;
