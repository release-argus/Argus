import { useQuery } from '@tanstack/react-query';
import { LoaderCircle } from 'lucide-react';
import type { ReactElement } from 'react';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableRow } from '@/components/ui/table';
import { useDelayedRender } from '@/hooks/use-delayed-render';
import { QUERY_KEYS } from '@/lib/query-keys';
import { cn } from '@/lib/utils';
import { mapRequest } from '@/utils/api/types/api-request-handler';

const titleMappings: Record<string, string> = {
	cwd: 'Working directory',
};
const ignoreCapitalise = new Set(['GOMAXPROCS', 'GOGC', 'GODEBUG']);

/**
 * @returns The status page, which includes tables of runtime info and of build info.
 */
export const Status = (): ReactElement => {
	const delayedRender = useDelayedRender(750);

	// Fetch the runtime info from the API.
	const { data: runtimeData } = useQuery({
		queryFn: () => mapRequest('STATUS_RUNTIME', null),
		queryKey: QUERY_KEYS.CONFIG.RUNTIME_INFO(),
		staleTime: Infinity, // won't change until Argus restarted.
	});
	// Fetch the build info from the API.
	const { data: buildData } = useQuery({
		queryFn: () => mapRequest('STATUS_BUILD', null),
		queryKey: QUERY_KEYS.CONFIG.BUILD_INFO(),
		staleTime: Infinity, // won't change until Argus restarted.
	});

	return (
		<div className="flex flex-col gap-8">
			<div>
				<h2 className="flex scroll-m-20 flex-row gap-2 pb-2 font-semibold text-3xl tracking-tight">
					Runtime Information
					{!runtimeData &&
						delayedRender(() => (
							<div className="h-8 items-center justify-center">
								<LoaderCircle className="h-full animate-spin" />
							</div>
						))}
				</h2>
				<Table className="border">
					<TableBody>
						{runtimeData
							? Object.entries(runtimeData).map(([k, v]) => {
									const title = (
										k in titleMappings ? titleMappings[k] : k
									).replaceAll('_', ' ');

									return (
										<TableRow className="odd:bg-muted/30" key={k}>
											<TableCell
												className={cn(
													'w-1/3 border-r py-4',
													!ignoreCapitalise.has(k) && 'capitalize',
												)}
											>
												{title}
											</TableCell>
											<TableCell>{v}</TableCell>
										</TableRow>
									);
								})
							: Array.from(new Array(4).keys()).map((num) => (
									<TableRow className="odd:bg-muted/30" key={num}>
										<TableCell className="w-1/3 border-r py-2">
											{delayedRender(() => (
												<Skeleton className="h-5 w-full" />
											))}
										</TableCell>
										<TableCell>
											{delayedRender(() => (
												<Skeleton className="h-5 w-full" />
											))}
										</TableCell>
									</TableRow>
								))}
					</TableBody>
				</Table>
			</div>
			<div>
				<h2 className="flex scroll-m-20 flex-row gap-2 pb-2 font-semibold text-3xl tracking-tight">
					Build Information
					{!buildData &&
						delayedRender(() => (
							<div className="h-8 items-center justify-center">
								<LoaderCircle className="h-full animate-spin" />
							</div>
						))}
				</h2>
				<Table className="border-x border-b">
					<TableBody>
						{buildData
							? Object.entries(buildData).map(([k, v]) => {
									const title = (
										k in titleMappings ? titleMappings[k] : k
									).replaceAll('_', ' ');

									return (
										<TableRow className="odd:bg-muted/30" key={k}>
											<TableCell
												className={cn(
													'border-r px-4 py-2 font-bold',
													!ignoreCapitalise.has(k) && 'capitalize',
												)}
											>
												{title}
											</TableCell>
											<TableCell className="px-4 py-2 font-bold">{v}</TableCell>
										</TableRow>
									);
								})
							: Array.from(new Array(3).keys()).map((num) => (
									<TableRow className="odd:bg-muted/30" key={num}>
										<TableCell className="flex w-1/3 flex-row border-r py-2">
											&nbsp;
										</TableCell>
										<TableCell>
											{delayedRender(() => (
												<Skeleton className="h-5 w-full" />
											))}
										</TableCell>
									</TableRow>
								))}
					</TableBody>
				</Table>
			</div>
		</div>
	);
};
