import type { Row } from '@tanstack/react-table';
import { Pencil } from 'lucide-react';
import { type FC, useCallback } from 'react';
import { useToolbar } from '@/components/approvals/toolbar/toolbar-context';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/contexts/auth';
import useModal from '@/hooks/use-modal';
import { DEPLOYED_VERSION_LOOKUP_TYPE } from '@/utils/api/types/config/service/deployed-version';
import type {
	ModalType,
	ServiceSummary,
} from '@/utils/api/types/config/summary';

type ServiceActionReleaseProps = {
	/* The row in the table */
	row: Row<ServiceSummary>;
};

/**
 * A functional component that provides action controls for interacting with a service's commands/webhooks,
 * and an 'edit' button when edit mode is enabled.
 *
 * @param row - The ServiceSummary data of the service.
 *
 * Features:
 * - Provides an 'Edit' action when edit mode is enabled.
 * - Displays 'Skip' button when an update is available and not yet skipped.
 * - Conditionally shows an 'Approve' or 'Resend' action for approving actions or retrying actions, based on service status and state.
 *
 * @returns A set of action controls for interacting with a service's commands/webhooks, and editing of that service.
 */
export const ServiceActionRelease: FC<ServiceActionReleaseProps> = ({
	row,
}) => {
	const { setModal } = useModal();
	const { hasPermission } = useAuth();
	const {
		values: { editMode },
	} = useToolbar();

	const service = row.original;
	// biome-ignore lint/correctness/useExhaustiveDependencies: setModal stable.
	const showModal = useCallback(
		(type: ModalType) => {
			setModal({ actionType: type, service: service });
		},
		[setModal],
	);

	const scope = { serviceID: service.id, tags: service.tags };
	const canEdit = hasPermission('service', 'update', scope);
	const canAction = hasPermission('service_action', 'execute', scope);

	const haveUpdateAction =
		(service.command ?? 0) > 0 || (service.webhook ?? 0) > 0;
	const updateAvailable = service.status?.state === 'AVAILABLE';
	const updateSkipped = service.status?.state === 'SKIPPED';
	const canApproveManually =
		!haveUpdateAction &&
		service.status?.state !== 'UP_TO_DATE' &&
		service.deployed_version_type === DEPLOYED_VERSION_LOOKUP_TYPE.MANUAL.value;

	return (
		<div className="flex flex-row items-center gap-x-2">
			{editMode && canEdit && (
				<div className="flex flex-row items-center gap-x-4">
					<Button
						aria-label="Edit service"
						className="size-6"
						onClick={() => showModal('EDIT')}
						size="sm"
						variant="secondary"
					>
						<Pencil />
					</Button>
				</div>
			)}
			{canAction && updateAvailable && !updateSkipped && (
				<Button
					aria-label="Reject release"
					key="reject"
					onClick={() => showModal('SKIP')}
					size="xs"
					variant="destructive"
				>
					Skip
				</Button>
			)}
			{canAction && haveUpdateAction && (
				<Button
					aria-label={updateSkipped ? 'Approve release' : 'Resend actions'}
					key="approve"
					onClick={() => showModal(updateAvailable ? 'SEND' : 'RESEND')}
					size="xs"
					variant={updateAvailable && !updateSkipped ? 'default' : 'secondary'}
				>
					{updateAvailable ? 'Approve' : 'Resend'}
				</Button>
			)}
			{canAction && canApproveManually && (
				<Button
					aria-label="Approve release"
					key="approve-manual"
					onClick={() => showModal('APPROVE_MANUAL')}
					size="xs"
					variant={updateSkipped ? 'secondary' : 'default'}
				>
					Approve
				</Button>
			)}
		</div>
	);
};
