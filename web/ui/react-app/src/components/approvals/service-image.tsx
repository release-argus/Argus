import { FC, useMemo } from 'react';
import {
	faCircleNotch,
	faGripVertical,
	faWindowMaximize,
} from '@fortawesome/free-solid-svg-icons';

import { Card } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { ServiceSummaryType } from 'types/summary';
import { faGithub } from '@fortawesome/free-brands-svg-icons';
import { useDelayedRender } from 'hooks/delayed-render';

interface Props {
	service_name: ServiceSummaryType['name'];
	service_type: ServiceSummaryType['type'];
	loading: ServiceSummaryType['loading'];
	icon?: ServiceSummaryType['icon'];
	icon_link_to?: ServiceSummaryType['icon_link_to'];
	visible: boolean;

	draggable?: boolean;
	dragHandleProps?: any; // Props from useSortable
}

/**
 * The service's image, with a loading spinner if the image is not loaded yet
 * and a link to the service. If the service has no icon, the service type icon (github/url) is displayed.
 *
 * @param service_name - The name of the service.
 * @param service_type - The type of the service.
 * @param loading - Whether the service is loading.
 * @param icon - The URL of the service's icon.
 * @param icon_link_to - The URL to link to when the icon is clicked.
 * @param visible - Whether the image should be visible.
 * @param draggable - Whether the service is draggable.
 * @param dragHandleProps - Props for the drag handle.
 * @returns A component that displays the image of the service.
 */
export const ServiceImage: FC<Props> = ({
	service_name,
	service_type,
	loading,
	icon,
	icon_link_to,
	visible,
	draggable = false,
	dragHandleProps,
}) => {
	const delayedRender = useDelayedRender(500);

	const imageStyles = {
		minWidth: 'fit-content',
		height: '6rem',
	};

	const iconRender = useMemo(() => {
		// URL icon.
		if (icon)
			return (
				<Card.Img
					variant="top"
					src={icon}
					alt={`${service_name}`}
					className="service-image"
				/>
			);

		// Loading spinner.
		if (loading)
			return (
				<div
					className="service-image"
					style={{ display: visible ? 'inline' : 'none' }}
				>
					{delayedRender(() => (
						<FontAwesomeIcon
							icon={faCircleNotch}
							style={{ ...imageStyles, padding: '0' }}
							className="service-image fa-spin"
						/>
					))}
				</div>
			);

		// Default icon.
		return (
			<FontAwesomeIcon
				icon={service_type === 'github' ? faGithub : faWindowMaximize}
				style={imageStyles}
				className="service-image"
			/>
		);
	}, [service_type, icon, loading, visible]);

	return (
		<div
			className="empty"
			style={{
				height: '7rem',
				display: visible ? 'flex' : 'none',
				position: 'relative',
			}}
		>
			{draggable && (
				<div
					{...dragHandleProps}
					style={{
						position: 'absolute',
						top: '0.5rem',
						left: '0.5rem',
						cursor: 'grab',
						padding: '0.5rem',
						touchAction: 'none',
						color: 'var(--bs-secondary-color)',
					}}
					aria-label="Drag handle"
				>
					<FontAwesomeIcon icon={faGripVertical} />
				</div>
			)}
			<a
				href={icon_link_to || undefined}
				target="_blank"
				rel="noreferrer noopener"
				style={{ color: 'inherit', display: 'contents' }}
			>
				{iconRender}
			</a>
		</div>
	);
};
