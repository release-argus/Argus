import { useQuery } from '@tanstack/react-query';
import { LoaderCircle } from 'lucide-react';
import type { ReactElement } from 'react';
import { useDelayedRender } from '@/hooks/use-delayed-render';
import { QUERY_KEYS } from '@/lib/query-keys';
import { mapRequest } from '@/utils/api/types/api-request-handler';

/**
 * @returns The configuration page, which includes a preformatted YAML object of the config.yml.
 */
export const Config = (): ReactElement => {
	const delayedRender = useDelayedRender(750);

	// Fetch the config YAML from the API.
	const { data, isFetching } = useQuery<string>({
		queryFn: () => mapRequest('CONFIG_GET', null),
		queryKey: QUERY_KEYS.CONFIG.RAW(),
		staleTime: 0,
	});

	return (
		<>
			<h2 className="flex scroll-m-20 flex-row gap-2 pb-2 font-semibold text-3xl tracking-tight">
				Configuration
				{isFetching &&
					delayedRender(() => (
						<div className="h-8 items-center justify-center">
							<LoaderCircle className="h-full animate-spin" />
						</div>
					))}
			</h2>
			{data && (
				<pre className="whitespace-pre-wrap bg-secondary p-4 font-mono text-sm">
					{data}
				</pre>
			)}
		</>
	);
};
