import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { useQuery } from '@tanstack/react-query';
import { GripVertical, Pencil } from 'lucide-react';
import { type FC, memo, useCallback, useMemo } from 'react';
import { Link } from 'react-router-dom';
import ServiceImage from '@/components/approvals/service-image';
import ServiceInfo from '@/components/approvals/service-info';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { useDelayedRender } from '@/hooks/use-delayed-render';
import useModal from '@/hooks/use-modal';
import { QUERY_KEYS } from '@/lib/query-keys';
import { cn } from '@/lib/utils';
import { isEmptyOrNull } from '@/utils';
import { mapRequest } from '@/utils/api/types/api-request-handler';
import type {
	ModalType,
	ServiceSummary,
} from '@/utils/api/types/config/summary';

type ServiceProps = {
	id: string;
	editable: boolean;
};

/**
 * A card with the service's information, including the service's image,
 * version info, and update info.
 *
 * @param id - The service ID.
 * @param editable - Whether edit mode is enabled.
 */
const Service: FC<ServiceProps> = ({ id, editable = false }) => {
	const delayedRender = useDelayedRender(250);
	const { setModal } = useModal();

	// Service summary.
	const { data } = useQuery({
		placeholderData: { id: id, loading: true },
		queryFn: () => mapRequest('SERVICE_SUMMARY', { serviceID: id }),
		queryKey: QUERY_KEYS.SERVICE.SUMMARY_ITEM(id),
		staleTime: Infinity,
	});

	// Sortable cards.
	const {
		attributes,
		listeners,
		setNodeRef,
		transform,
		transition,
		isDragging,
	} = useSortable({
		disabled: !editable,
		id,
	});

	// biome-ignore lint/correctness/useExhaustiveDependencies: setModal stable.
	const showModal = useCallback(
		(type: ModalType, service: ServiceSummary) =>
			setModal({ actionType: type, service: service }),
		[],
	);

	const updateStatus = useMemo(() => {
		const updateAvailable = data?.status?.state === 'AVAILABLE';
		const updateSkipped = data?.status?.state === 'SKIPPED';
		const updateWarning = updateAvailable && !updateSkipped;

		return {
			// Update available when both 'latest' and 'deployed' versions defined, and differ.
			available: updateAvailable,
			// className for possible warning state
			className: cn(
				updateWarning && [
					'text-foreground',
					data?.active !== false && 'bg-accent',
				],
			),
			hasActions:
				(data?.command ?? 0) > 0 ||
				(data?.webhook ?? 0) > 0 ||
				updateAvailable ||
				updateSkipped,
			// Version not found.
			not_found:
				isEmptyOrNull(data?.status?.deployed_version) ||
				isEmptyOrNull(data?.status?.latest_version) ||
				isEmptyOrNull(data?.status?.last_queried),
			// Update available, and 'approved' version is a skip of the 'latest'.
			skipped: updateSkipped,
			// 'New' version found (and not skipped).
			warning: updateWarning,
		};
	}, [data]);

	const dragStyle = {
		transform: CSS.Transform.toString(transform),
		transition,
	};

	return (
		<Card
			className={cn(
				'gap-0 shadow',
				isDragging && 'z-100 border-primary/50 bg-secondary',
				updateStatus.not_found
					? delayedRender(() => updateStatus.className, 'default')
					: updateStatus.className,
			)}
			key={data?.id}
			ref={setNodeRef}
			style={dragStyle}
		>
			<CardHeader
				className="relative flex h-full flex-col items-center text-balance text-center"
				key={`${data?.id}-title`}
			>
				{data?.url ? (
					<Button
						asChild
						className="m-auto h-min w-fit items-start justify-start p-0 text-foreground"
						variant="link"
					>
						<Link rel="noreferrer noopener" target="_blank" to={data?.url}>
							<h4 className="whitespace-normal font-semibold text-xl tracking-tight">
								{data?.name ?? data?.id}
							</h4>
						</Link>
					</Button>
				) : (
					<h4 className="my-auto cursor-default text-center font-semibold text-xl tracking-tight">
						{data?.name ?? data?.id}
					</h4>
				)}
				{editable && (
					<div className="-top-2.5 absolute right-0.5 z-1 flex flex-col">
						<Button
							aria-label="Edit service"
							className="size-6"
							onClick={() => data && showModal('EDIT', data)}
							size="sm"
							variant="secondary"
						>
							<Pencil />
						</Button>
						<Button
							{...listeners}
							{...attributes}
							aria-label="Drag handle"
							className="cursor-grab touch-none px-0! py-2 text-muted-foreground"
							size="sm"
							variant="ghost"
						>
							<GripVertical />
						</Button>
					</div>
				)}
			</CardHeader>

			<CardContent className="px-2">
				<div
					className={cn(
						'flex gap-4 border-0 p-2',
						data?.active === false &&
							'border-2 border-[var(--muted-foreground)] bg-[repeating-linear-gradient(45deg,var(--muted)_0px,var(--muted)_20px,var(--muted-foreground)_20px,var(--muted-foreground)_40px)] p-0.5',
					)}
				>
					<ServiceImage service={data} />
					<ServiceInfo
						service={data}
						updateAvailable={updateStatus.available}
						updateSkipped={updateStatus.skipped}
					/>
				</div>
			</CardContent>
		</Card>
	);
};

export default memo(Service);
