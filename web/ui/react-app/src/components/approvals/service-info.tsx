import {
	Button,
	Card,
	Container,
	ListGroup,
	OverlayTrigger,
	Tooltip,
} from 'react-bootstrap';
import { FC, useCallback, useContext, useMemo } from 'react';
import { ModalType, ServiceSummaryType } from 'types/summary';
import {
	faArrowRotateRight,
	faCheck,
	faInfo,
	faInfoCircle,
	faSatelliteDish,
	faTimes,
} from '@fortawesome/free-solid-svg-icons';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { ModalContext } from 'contexts/modal';
import { formatRelative } from 'date-fns';
import { isEmptyOrNull } from 'utils';
import { useDelayedRender } from 'hooks/delayed-render';

interface Props {
	service: ServiceSummaryType;
	updateAvailable: boolean;
	updateSkipped: boolean;
	toggleShowUpdateInfo: () => void;
}

/**
 * The service's information, including the latest version, the deployed version,
 * the last time the service was queried, and the update options if latest and deployed versions differ.
 *
 * @param service - The service the information belongs to.
 * @param updateAvailable - Whether an update is available.
 * @param updateSkipped - Whether the update has been skipped.
 * @param toggleShowUpdateInfo - Function to toggle showing the update information.
 * @returns A component that displays the service's version information.
 */
export const ServiceInfo: FC<Props> = ({
	service,
	updateAvailable,
	updateSkipped,
	toggleShowUpdateInfo,
}) => {
	const { handleModal } = useContext(ModalContext);
	const delayedRender = useDelayedRender(250);

	const showModal = useCallback(
		(type: ModalType, service: ServiceSummaryType) => {
			handleModal(type, service);
		},
		[handleModal],
	);

	const status = useMemo(
		() => ({
			// If the version hasn't been found.
			not_found:
				isEmptyOrNull(service.status?.deployed_version) ||
				isEmptyOrNull(service.status?.latest_version) ||
				isEmptyOrNull(service.status?.last_queried),
			// If a new version has been found (and not skipped).
			warning: updateAvailable && !updateSkipped,
			// If the latest version is the same as the approved version.
			updateApproved:
				service?.status?.latest_version !== undefined &&
				service.status.latest_version === service?.status?.approved_version,
		}),
		[service, updateAvailable, updateSkipped],
	);

	const deployedVersionIcon = service.has_deployed_version ? (
		<OverlayTrigger
			key="deployed-service"
			placement="top"
			delay={{ show: 500, hide: 500 }}
			overlay={
				<Tooltip id="tooltip-deployed-service">
					of the deployed {service.id}
				</Tooltip>
			}
		>
			<FontAwesomeIcon
				style={{ paddingLeft: '0.5rem', paddingBottom: '0.1rem' }}
				icon={faSatelliteDish}
				aria-label="Monitoring the deployed service's version"
			/>
		</OverlayTrigger>
	) : null;

	const skippedVersionIcon =
		updateSkipped && service.status?.approved_version ? (
			<OverlayTrigger
				key="skipped-version"
				placement="top"
				delay={{ show: 500, hide: 500 }}
				overlay={
					<Tooltip id="tooltip-skipped-version">
						Skipped {service.status.approved_version.slice('SKIP_'.length)}
					</Tooltip>
				}
			>
				<FontAwesomeIcon
					icon={faInfoCircle}
					style={{ paddingLeft: '0.5rem', paddingBottom: '0.1rem' }}
					aria-describedby="tooltip-skipped-version"
				/>
			</OverlayTrigger>
		) : null;

	const actionType = service.webhook ? 'WebHooks' : 'Commands';

	const actionReleaseButton =
		(service.webhook || service.command) &&
		(!updateAvailable || updateSkipped) ? (
			<OverlayTrigger
				key="resend"
				placement="top"
				delay={{ show: 500, hide: 500 }}
				overlay={
					<Tooltip id="tooltip-resend">
						{updateSkipped
							? 'Approve this release'
							: `Resend the ${actionType}`}
					</Tooltip>
				}
			>
				<Button
					variant="secondary"
					size="sm"
					onClick={() => showModal(updateSkipped ? 'SEND' : 'RESEND', service)}
					disabled={service.loading || service.active === false}
					aria-describedby="tooltip-resend"
				>
					<FontAwesomeIcon
						icon={updateSkipped ? faCheck : faArrowRotateRight}
					/>
				</Button>
			</OverlayTrigger>
		) : null;

	return (
		<Container
			style={{
				padding: '0px',
			}}
			className={
				status.not_found
					? delayedRender(() => 'service-warning rounded-bottom', 'default')
					: status.warning
					? 'service-warning rounded-bottom'
					: 'default'
			}
		>
			<ListGroup className="list-group-flush">
				{updateAvailable && !updateSkipped ? (
					<>
						<ListGroup.Item
							key="update-available"
							className={'service-item update-options'}
							style={{ color: 'inherit' }}
							variant="secondary"
						>
							{status.updateApproved && (service.webhook || service.command)
								? `${actionType} already sent:`
								: 'Update available:'}
						</ListGroup.Item>
						<ListGroup.Item
							key="update-buttons"
							className={'service-item update-options'}
							variant="secondary"
							style={{ paddingTop: '0.25rem' }}
						>
							<Button
								key="details"
								className="btn-flex btn-update-action"
								variant="primary"
								onClick={toggleShowUpdateInfo}
								aria-label="Show update details"
							>
								<FontAwesomeIcon icon={faInfo} />
							</Button>
							<Button
								key="approve"
								className={`btn-flex btn-update-action${
									service.webhook || service.command ? '' : ' hiddenElement'
								}`}
								variant="success"
								onClick={() =>
									showModal(status.updateApproved ? 'RESEND' : 'SEND', service)
								}
								disabled={!(service.webhook || service.command)}
								aria-label="Approve release"
							>
								<FontAwesomeIcon icon={faCheck} />
							</Button>
							<Button
								key="reject"
								className="btn-flex btn-update-action"
								variant="danger"
								onClick={() =>
									showModal(
										service.webhook || service.command ? 'SKIP' : 'SKIP_NO_WH',
										service,
									)
								}
								aria-label="Reject release"
							>
								<FontAwesomeIcon icon={faTimes} color="white" />
							</Button>
						</ListGroup.Item>
					</>
				) : (
					<ListGroup.Item
						key="deployed_v"
						variant={
							status.not_found
								? delayedRender(() => 'warning', 'secondary')
								: status.warning
								? 'warning'
								: 'secondary'
						}
						className={
							'service-item' +
							(service.webhook || service.command ? '' : ' justify-left')
						}
					>
						<div style={{ margin: 0 }}>
							Current version:
							{deployedVersionIcon}
							{skippedVersionIcon}
							<br />
							<div style={{ display: 'flex', margin: 0 }}>
								<OverlayTrigger
									key="deployed-version"
									placement="top"
									delay={{ show: 500, hide: 500 }}
									overlay={
										service?.status?.deployed_version_timestamp ? (
											<Tooltip id="tooltip-deployed-version">
												<>
													{formatRelative(
														new Date(service.status.deployed_version_timestamp),
														new Date(),
													)}
												</>
											</Tooltip>
										) : (
											<>Unknown</>
										)
									}
								>
									<p
										style={{ margin: 0 }}
										aria-describedby="tooltip-deployed-version"
									>
										{service?.status?.deployed_version
											? service.status.deployed_version
											: 'Unknown'}{' '}
									</p>
								</OverlayTrigger>
							</div>
						</div>
						{actionReleaseButton}
					</ListGroup.Item>
				)}
			</ListGroup>
			<Card.Footer>
				<small
					className={
						'text-muted' +
						(status.not_found
							? delayedRender(() => ' service-warning rounded-bottom', '')
							: status.warning
							? ' service-warning rounded-bottom'
							: '')
					}
				>
					{service?.status?.last_queried ? (
						<>
							queried{' '}
							{formatRelative(
								new Date(service.status.last_queried),
								new Date(),
							)}
						</>
					) : service.loading ? (
						'loading'
					) : (
						'no successful queries'
					)}
				</small>
			</Card.Footer>
		</Container>
	);
};
