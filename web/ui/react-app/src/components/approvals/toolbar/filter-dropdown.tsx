import { CheckIcon, Eye } from 'lucide-react';
import { type FC, useCallback } from 'react';
import { useToolbar } from '@/components/approvals/toolbar/toolbar-context';
import { Button } from '@/components/ui/button';
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { optionStyles } from '@/components/ui/react-select/helper';
import Tip from '@/components/ui/tip';
import {
	ACTIVE_HIDE_VALUES,
	DEFAULT_HIDE_VALUE,
	HideValue,
	type HideValueType,
	toolbarHideOptions,
} from '@/constants/toolbar';
import { cn } from '@/lib/utils';

type HideOptionKey = (typeof toolbarHideOptions)[number]['key'];

/**
 * FilterDropdown
 *
 * A dropdown component for toggling visibility filters on services.
 * It lists all `HIDE_OPTIONS`, allowing users to show or hide specific statuses,
 * and includes a reset option to restore default visibility (`DEFAULT_HIDE_VALUE`).
 */
const FilterDropdown: FC = () => {
	const { values, setHide } = useToolbar();
	const currentValues = values.hide;

	const handleOptionClick = useCallback(
		(key: HideOptionKey) => {
			const option = toolbarHideOptions.find((opt) => opt.key === key);
			if (!option) return;
			const { value: clickedValue } = option;

			const toggle = (val: number) => {
				const newValues = currentValues.includes(val as HideValueType)
					? currentValues.filter((v) => v !== val)
					: [...currentValues, val];
				setHide(newValues);
			};

			const flipActiveValues = () => {
				const newActiveHidden = ACTIVE_HIDE_VALUES.filter(
					(v) => !currentValues.includes(v),
				);
				const inactiveIsHidden = currentValues.includes(HideValue.Inactive);
				const newValues = inactiveIsHidden
					? [...newActiveHidden, HideValue.Inactive]
					: newActiveHidden;
				setHide(newValues);
			};

			if (ACTIVE_HIDE_VALUES.includes(clickedValue)) {
				const otherActiveValues = ACTIVE_HIDE_VALUES.filter(
					(v) => v !== clickedValue,
				);
				// If all other active statuses hidden, flip them all.
				if (otherActiveValues.every((v) => currentValues.includes(v))) {
					flipActiveValues();
					return;
				}
			}

			toggle(clickedValue);
		},
		[currentValues, setHide],
	);

	const handleReset = useCallback(() => {
		setHide(DEFAULT_HIDE_VALUE);
	}, [setHide]);

	const filterButtonTooltip = 'Filter shown services';

	return (
		<DropdownMenu>
			<DropdownMenuTrigger asChild>
				<div className="h-full cursor-pointer">
					<Tip
						content={filterButtonTooltip}
						delayDuration={500}
						touchDelayDuration={250}
					>
						<Button
							aria-label={filterButtonTooltip}
							className="rounded-e-none"
							size="icon-md"
							variant="outline"
						>
							<Eye />
						</Button>
					</Tip>
				</div>
			</DropdownMenuTrigger>
			<DropdownMenuContent className="w-max">
				{toolbarHideOptions.map(({ key, label, value }) => {
					const isSelected = currentValues.includes(value);
					return (
						<DropdownMenuItem
							className={cn(
								'cursor-pointer',
								optionStyles.base,
								isSelected && optionStyles.selected,
							)}
							key={key}
							onClick={() => handleOptionClick(key)}
						>
							<div className="flex w-30 flex-row justify-between sm:w-36">
								<span>{label}</span>
								{isSelected && (
									<span className="flex h-full items-center justify-center">
										<CheckIcon className="size-4" />
									</span>
								)}
							</div>
						</DropdownMenuItem>
					);
				})}
				<DropdownMenuItem className="cursor-pointer" onClick={handleReset}>
					Reset
				</DropdownMenuItem>
			</DropdownMenuContent>
		</DropdownMenu>
	);
};

export default FilterDropdown;
