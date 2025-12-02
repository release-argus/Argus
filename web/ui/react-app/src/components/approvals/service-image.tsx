import { SiGithub } from '@icons-pack/react-simple-icons';
import { AppWindow, LoaderCircle } from 'lucide-react';
import { type FC, useMemo } from 'react';
import { useDelayedRender } from '@/hooks/use-delayed-render';
import { LATEST_VERSION_LOOKUP_TYPE } from '@/utils/api/types/config/service/latest-version';
import type { ServiceSummary } from '@/utils/api/types/config/summary';

type ServiceImageProps = {
	service?: ServiceSummary;
};

/**
 * The service's image, with a possible loading spinner and a link to the service.
 * If the service has no icon, the service type icon (github/url) is displayed.
 *
 * @param service - The service.
 * @returns The image tied to the service.
 */
const ServiceImage: FC<ServiceImageProps> = ({ service }) => {
	const delayedRender = useDelayedRender(500);
	const {
		type: serviceType,
		icon: icon,
		icon_link_to: iconLinkTo,
		loading,
	} = service ?? {};

	// biome-ignore lint/correctness/useExhaustiveDependencies: delayedRender stable.
	const iconRender = useMemo(() => {
		// URL icon.
		if (icon)
			return <img alt="" className="!size-full object-contain" src={icon} />;

		// Loading spinner.
		if (loading)
			return (
				<div className="inline">
					{delayedRender(() => (
						<LoaderCircle className="size-full animate-spin text-muted-foreground" />
					))}
				</div>
			);

		// Default icon.
		const ServiceIcon = LATEST_VERSION_LOOKUP_TYPE.GITHUB.value
			? SiGithub
			: AppWindow;
		return <ServiceIcon className="!size-full object-contain" />;
	}, [serviceType, icon, loading]);

	return (
		<div className="relative my-auto flex aspect-[3/2] size-22 items-center justify-center">
			<a
				className="flex size-full items-center justify-center"
				href={iconLinkTo || undefined}
				rel="noreferrer noopener"
				target="_blank"
			>
				{iconRender}
			</a>
		</div>
	);
};

export default ServiceImage;
