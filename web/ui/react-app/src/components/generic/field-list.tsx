import { Minus, Plus } from 'lucide-react';
import { type FC, useCallback, useEffect, useMemo } from 'react';
import { useFieldArray, useFormContext, useWatch } from 'react-hook-form';
import { FieldLabel, FieldText } from '@/components/generic/field';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import { FieldGroup, FieldSet } from '@/components/ui/field';
import type { StringFieldArray } from '@/types/util';
import { isEmptyArray } from '@/utils';
import { isUsingDefaults } from '@/utils/api/types/config-edit/validators';

type BaseProps = {
	/* The name of the field. */
	name: string;
	/* The form label to display. */
	label?: string;

	/* The default values for the field. */
	defaults?: StringFieldArray;
};

type FieldListProps = BaseProps & {
	/* The tooltip on the field label. */
	tooltip?: TooltipWithAriaProps;
};

/**
 * A labelled set of form fields for a list of strings.
 *
 * @param name - The name of the field in the form.
 * @param label - The label for the field.
 * @param tooltip - The tooltip on the field label.
 * @param tooltip.type - The tooltip content type: either 'string' for plain text or 'element' for a React element.
 * @param tooltip.side - The wide to render the tooltip content.
 * @param tooltip.size - The size of the tooltip.
 * @param tooltip.delayDuration - Time before rendering the tooltip.
 * @param defaults - The default values for the field.
 */
const FieldList: FC<FieldListProps> = ({
	name,
	label = 'List',
	tooltip,

	defaults,
}) => {
	const { setValue, trigger } = useFormContext();
	const { fields, append, remove } = useFieldArray({ name: name });
	// biome-ignore lint/correctness/useExhaustiveDependencies: append stable.
	const addItem = useCallback(() => {
		append({ arg: '' }, { shouldFocus: false });
	}, []);

	// Keep track of the array values, so we can use defaults when empty.
	const fieldValues = useWatch({ name: name }) as StringFieldArray | undefined;
	// Use defaults when fieldValues undefined or the same as the defaults.
	const usingDefaults = useMemo(
		() => isUsingDefaults({ arg: fieldValues, defaultValue: defaults }),
		[fieldValues, defaults],
	);
	// Trigger validation on change of defaults used/not.
	// biome-ignore lint/correctness/useExhaustiveDependencies: defaults covered by usingDefaults.
	useEffect(() => {
		void trigger(name);

		// Give defaults back when field empty.
		if (defaults && isEmptyArray(fieldValues)) {
			const defaultsHollow = defaults.map(() => ({ arg: '' }));
			setValue(name, defaultsHollow);
		}
	}, [usingDefaults]);

	const placeholder = useCallback(
		(index: number) => (usingDefaults ? (defaults?.[index]?.arg ?? '') : ''),
		[usingDefaults, defaults],
	);

	// On load, ensure we don't have another type's actions
	// and give the defaults if not overridden.
	// biome-ignore lint/correctness/useExhaustiveDependencies: on-load only.
	useEffect(() => {
		if (!fieldValues) return;

		for (const item of fieldValues) {
			const keys = Object.keys(item);
			if (keys.length !== 1 || !keys.includes('arg')) {
				const defaultsHollow = (defaults ?? []).map(() => ({ arg: '' }));
				setValue(name, defaultsHollow);
				break;
			}
		}
	}, []);

	// Remove the last item when not the only one, or it doesn't match the defaults.
	// biome-ignore lint/correctness/useExhaustiveDependencies: remove stable.
	const removeLast = useCallback(() => {
		if (!(usingDefaults && fields.length == 1)) remove(fields.length - 1);
	}, [fields.length, usingDefaults]);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid space-y-1">
			<FieldGroup className="col-span-full flex w-full flex-row items-center justify-between">
				<FieldLabel text={label} tooltip={tooltip} />
				<ButtonGroup>
					<Button
						aria-label={`Add new ${label}`}
						onClick={addItem}
						size="icon-xs"
						variant="ghost"
					>
						<Plus />
					</Button>
					<Button
						aria-label={`Remove last ${label}`}
						disabled={fields.length === 0}
						onClick={removeLast}
						size="icon-xs"
						variant="ghost"
					>
						<Minus />
					</Button>
				</ButtonGroup>
			</FieldGroup>
			<FieldGroup className="col-span-full grid grid-cols-subgrid gap-2">
				{fields.map(({ id }, index) => (
					<FieldText
						className="grid"
						defaultVal={placeholder(index)}
						key={id}
						name={`${name}.${index}.arg`}
						required
					/>
				))}
			</FieldGroup>
		</FieldSet>
	);
};

export default FieldList;
