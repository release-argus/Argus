import { Minus, Plus } from 'lucide-react';
import { type FC, memo, useCallback, useEffect, useMemo } from 'react';
import { useFieldArray, useFormContext, useWatch } from 'react-hook-form';
import FieldKeyVal from '@/components/generic/field-key-val';
import FieldLabelWithTooltip from '@/components/generic/field-label';
import type { HeaderPlaceholders } from '@/components/generic/field-shared';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import { cn } from '@/lib/utils';
import { isEmptyArray } from '@/utils';
import type { CustomHeader } from '@/utils/api/types/config/shared';
import { isUsingDefaults } from '@/utils/api/types/config-edit/validators';

type BaseProps = {
	/* The name of the field. */
	name: string;
	/* The form label to display. */
	label?: string;
	/* The column width on XS+ screens. */
	colSpan?: number;
	/* Optional placeholders for the key and value fields. */
	placeholders?: HeaderPlaceholders;

	/* The default values for the field. */
	defaults?: CustomHeader[];
};

type FieldKeyValMapProps = BaseProps & {
	/* The tooltip on the field label. */
	tooltip?: TooltipWithAriaProps;
};

/**
 * The labelled form fields for a key-value map.
 *
 * @param name - The name of the field in the form.
 * @param label - The label for the field.
 * @param colSpan - The width of the field.
 * @param tooltip - The tooltip on the field label.
 * @param tooltip.type - The tooltip content type: either 'string' for plain text or 'element' for a React element.
 * @param tooltip.side - The wide to render the tooltip content.
 * @param tooltip.size - The size of the tooltip.
 * @param tooltip.delayDuration - Time before rendering the tooltip.
 * @param placeholders - Optional placeholders for the key/value fields.
 * @param defaults - The default values for the field.
 */
const FieldKeyValMap: FC<FieldKeyValMapProps> = ({
	name,
	label = 'Headers',
	colSpan = 12,
	tooltip,
	placeholders,

	defaults,
}) => {
	const { trigger, setValue } = useFormContext();
	const { fields, append, remove } = useFieldArray({ name: name });

	// biome-ignore lint/correctness/useExhaustiveDependencies: append stable.
	const addItem = useCallback(() => {
		append({ key: '', value: '' }, { shouldFocus: false });
	}, []);

	// Keep track of the array values, so we can use defaults when empty.
	const fieldValues = useWatch({ name: name }) as CustomHeader[];
	// Use defaults when fieldValues undefined or the same as the defaults.
	const usingDefaults = useMemo(
		() => isUsingDefaults({ arg: fieldValues, defaultValue: defaults }),
		[fieldValues, defaults],
	);
	// Trigger validation on change of defaults used/not.
	// Reset to defaults when empty.
	// biome-ignore lint/correctness/useExhaustiveDependencies: usingDefaults covers fieldValues.
	useEffect(() => {
		void trigger(name);

		// Give defaults back when field empty.
		if (defaults && isEmptyArray(fieldValues)) {
			const defaultsHollow = defaults.map(() => ({ key: '', value: '' }));
			setValue(name, defaultsHollow);
		}
	}, [usingDefaults, defaults]);

	// Remove item at given index.
	// biome-ignore lint/correctness/useExhaustiveDependencies: name stable.
	const removeItem = useCallback(
		(index: number) => () => {
			document.getElementById(name)?.focus();
			remove(index);
		},
		[],
	);
	// Remove the last item if not the only one or doesn't match the defaults.
	// biome-ignore lint/correctness/useExhaustiveDependencies: removeItem stable.
	const removeLast = useCallback(
		() => () => {
			if (!(usingDefaults && fields.length === 1)) removeItem(0)();
		},
		[fields.length, usingDefaults],
	);

	return (
		<div className="col-span-full grid grid-cols-subgrid">
			<div className="col-span-full flex w-full items-center justify-between">
				<FieldLabelWithTooltip text={label} tooltip={tooltip} />
				<ButtonGroup>
					<Button
						aria-label={`Add new ${label}`}
						id={name}
						onClick={addItem}
						size="icon-xs"
						variant="ghost"
					>
						<Plus />
					</Button>
					<Button
						aria-label={`Remove last ${label}`}
						disabled={isEmptyArray(fields)}
						onClick={removeLast()}
						size="icon-xs"
						variant="ghost"
					>
						<Minus />
					</Button>
				</ButtonGroup>
			</div>
			<div
				className={cn('grid grid-cols-subgrid gap-2', `col-span-${colSpan}`)}
			>
				{fields.map(({ id: _id }, index) => (
					<FieldKeyVal
						colSpan={colSpan - 1}
						defaults={usingDefaults ? defaults?.[index] : undefined}
						key={_id}
						name={`${name}.${index}`}
						placeholders={placeholders}
						removeMe={
							// Disable 'remove' if one item, and it matches the defaults.
							fieldValues.length === 1 ? removeLast() : removeItem(index)
						}
					/>
				))}
			</div>
		</div>
	);
};

export default memo(FieldKeyValMap);
