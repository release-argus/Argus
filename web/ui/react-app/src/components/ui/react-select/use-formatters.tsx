// SOURCE: https://gist.github.com/ilkou/7bf2dbd42a7faf70053b43034fc4b5a4
import { clsx } from 'clsx';
import type { ReactNode } from 'react';
import type { GroupBase } from 'react-select';

/**
 * This hook could be added to your select component if needed:
 *   const formatters = useFormatters()
 *   <Select
 *     // other props
 *     {...formatters}
 *   />
 */
export const useFormatters = () => {
	// useful for CreatableSelect
	const formatCreateLabel = (label: unknown): ReactNode => (
		<span className={clsx('text-sm')}>
			Add <span className={clsx('font-semibold')}>{`"${String(label)}"`}</span>
		</span>
	);

	// useful for GroupedOptions
	const formatGroupLabel: (group: GroupBase<unknown>) => ReactNode = (data) => (
		<div className={'flex items-center justify-between'}>
			<span>{data.label}</span>
			<span
				className={
					'rounded-md bg-secondary px-1 font-normal text-secondary-foreground text-xs shadow-sm'
				}
			>
				{data.options.length}
			</span>
		</div>
	);
	return {
		formatCreateLabel,
		formatGroupLabel,
	};
};
