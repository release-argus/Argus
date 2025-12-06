import { useQuery } from '@tanstack/react-query';
import { LoaderCircle } from 'lucide-react';
import type { ReactElement } from 'react';
import { Skeleton } from '@/components/ui/skeleton';
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from '@/components/ui/table';
import { useDelayedRender } from '@/hooks/use-delayed-render';
import { QUERY_KEYS } from '@/lib/query-keys';
import { mapRequest } from '@/utils/api/types/api-request-handler';

/**
 * @returns The command-line flags page, which includes a table of all the command-line flags.
 */
export const Flags = (): ReactElement => {
	const delayedRender = useDelayedRender(750);

	// Fetch the command-line flags from the API.
	const { data, isSuccess } = useQuery({
		queryFn: () => mapRequest('CONFIG_FLAGS', null),
		queryKey: QUERY_KEYS.CONFIG.CLI_FLAGS(),
		staleTime: Infinity,
	});

	return (
		<>
			<h2 className="flex scroll-m-20 flex-row gap-2 pb-2 font-semibold text-3xl tracking-tight">
				Command-Line Flags
				{!isSuccess &&
					delayedRender(() => (
						<div className="h-8 items-center justify-center">
							<LoaderCircle className="h-full animate-spin" />
						</div>
					))}
			</h2>
			<Table className="border">
				<TableHeader>
					<TableRow>
						<TableHead className="w-1/3 border-r px-4 py-2 font-bold">
							Flag
						</TableHead>
						<TableHead className="px-4 py-2 font-bold">Value</TableHead>
					</TableRow>
				</TableHeader>
				<TableBody>
					{isSuccess
						? Object.entries(data).map(([k, v]) => (
								<TableRow className="odd:bg-muted/30" key={k}>
									<TableCell className="border-r px-4 py-2 font-bold">{`-${k}`}</TableCell>
									<TableCell className="px-4 py-2 font-bold">
										{v == undefined ? '' : v.toString()}
									</TableCell>
								</TableRow>
							))
						: Array.from(new Array(9).keys()).map((num) => (
								<TableRow className="odd:bg-muted/30" key={num}>
									<TableCell className="border-r py-3">
										{delayedRender(() => (
											<Skeleton className="h-4 w-full" />
										))}
									</TableCell>
									<TableCell>
										{delayedRender(() => (
											<Skeleton className="h-4 w-full" />
										))}
									</TableCell>
								</TableRow>
							))}
				</TableBody>
			</Table>
		</>
	);
};
