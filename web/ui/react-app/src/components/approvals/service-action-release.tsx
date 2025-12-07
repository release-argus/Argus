import { type FC, useCallback, useMemo } from 'react';
import { Button } from '@/components/ui/button';
import useModal from '@/hooks/use-modal';
import type {
	ModalType,
	ServiceSummary,
} from '@/utils/api/types/config/summary';

type ServiceActionReleaseProps = {
	service: ServiceSummary;
	updateAvailable: boolean;
	updateSkipped: boolean;
};

/**
 * Displays possible actions for service releases.
 * Either approve, resend or skip based on the service's state
 * and the presence of WebHooks or Commands for this service.
 *
 * @param service - The service to display actions for.
 * @param updateAvailable - Known update available for this service.
 * @param updateSkipped - Latest known release skipped.
 */
const ServiceActionRelease: FC<ServiceActionReleaseProps> = ({
	service,
	updateAvailable,
	updateSkipped,
}) => {
	const { setModal } = useModal();

	const showModal = useCallback(
		(type: ModalType, service: ServiceSummary) => {
			setModal({ actionType: type, service: service });
		},
		[setModal],
	);

	const info = useMemo(() => {
		let actionType: string | null = null;
		if (service.webhook) {
			if (service.command) {
				actionType = 'Commands/WebHooks';
			} else {
				actionType = 'WebHooks';
			}
		} else if (service.command) {
			actionType = 'Commands';
		}

		const haveUpdateAction = service.webhook || service.command;

		return {
			actionType,
			haveUpdateAction,
		};
	}, [service]);

	return (
		<div className="flex flex-row items-center gap-x-2">
			{updateAvailable && !updateSkipped && (
				<Button
					aria-label="Reject release"
					key="reject"
					onClick={() =>
						showModal(info.haveUpdateAction ? 'SKIP' : 'SKIP_NO_WH', service)
					}
					size="xs"
					variant="destructive"
				>
					Skip
				</Button>
			)}
			{info.haveUpdateAction && (
				<Button
					aria-label={updateSkipped ? 'Approve release' : 'Resend actions'}
					key="approve"
					onClick={() =>
						showModal(updateAvailable ? 'SEND' : 'RESEND', service)
					}
					size="xs"
					variant={updateAvailable && !updateSkipped ? 'default' : 'secondary'}
				>
					{updateAvailable ? 'Approve' : 'Resend'}
				</Button>
			)}
		</div>
	);
};

export default ServiceActionRelease;
