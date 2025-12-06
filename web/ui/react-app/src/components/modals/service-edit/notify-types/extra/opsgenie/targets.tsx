import { Minus, Plus } from 'lucide-react';
import { type FC, memo, useCallback, useEffect, useMemo } from 'react';
import { useFieldArray, useFormContext, useWatch } from 'react-hook-form';
import { FieldLabel } from '@/components/generic/field';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import OpsGenieTarget from '@/components/modals/service-edit/notify-types/extra/opsgenie/target';
import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import { isEmptyArray } from '@/utils';
import type { OpsGenieTargetsSchema } from '@/utils/api/types/config-edit/notify/types/opsgenie';
import { isUsingDefaults } from '@/utils/api/types/config-edit/validators';

type BaseProps = {
	/* The name of the field in the form. */
	name: string;
	/* The label for the field. */
	label: string;

	/* The default values for the field. */
	defaults?: OpsGenieTargetsSchema;
};

type OpsGenieTargetsProps = BaseProps & {
	/* The tooltip on the field label. */
	tooltip: TooltipWithAriaProps;
};

/**
 * OpsGenieTargets returns the form fields for a list of OpsGenie targets.
 *
 * @param name - The name of the field in the form.
 * @param label - The label for the field.
 * @param tooltip - The tooltip on the field label.
 * @param tooltip.type - 'string' | 'element'.
 * @param tooltip.side - The wide to render the tooltip content.
 * @param tooltip.size - The size of the tooltip.
 * @param tooltip.delayDuration - Time before rendering the tooltip.
 * @param defaults - The default values for the field.
 * @returns A set of form fields for a list of OpsGenie targets.
 */
const OpsGenieTargets: FC<OpsGenieTargetsProps> = ({
	name,
	label,
	tooltip,
	defaults,
}) => {
	const { setValue, trigger } = useFormContext();
	const { fields, append, remove } = useFieldArray({ name: name });

	// biome-ignore lint/correctness/useExhaustiveDependencies: append stable.
	const addItem = useCallback(() => {
		append(
			{
				sub_type: 'id',
				type: 'team',
				value: '',
			},
			{ shouldFocus: false },
		);
	}, []);

	// Keep track of the array values so can switch to defaults when unchanged.
	// @ts-ignore: control in context.
	const fieldValues = useWatch({ name: name }) as OpsGenieTargetsSchema;
	// usingDefaults when fieldValues undefined, or match defaults.
	const usingDefaults = useMemo(
		() =>
			isUsingDefaults({
				arg: fieldValues,
				defaultValue: defaults,
				matchingFieldsEndsWiths: ['.type', '.sub_type'],
			}),
		[fieldValues, defaults],
	);
	// Validate on change of defaults being usable.
	// Give defaults back when empty.
	// biome-ignore lint/correctness/useExhaustiveDependencies: defaults stable.
	useEffect(() => {
		trigger(name).catch(() => {
			return;
		});

		// Give defaults back if field empty.
		if (defaults && isEmptyArray(fieldValues))
			setValue(
				name,
				defaults.map((t) => ({
					sub_type: t.sub_type,
					type: t.type,
				})),
			);
	}, [usingDefaults]);

	// Remove the last item if not the only one, or doesn't match the defaults.
	// biome-ignore lint/correctness/useExhaustiveDependencies: remove stable.
	const removeLast = useCallback(() => {
		if (!(usingDefaults && fields.length == 1)) remove(fields.length - 1);
	}, [fields.length, usingDefaults]);

	return (
		<div className="col-span-full grid grid-cols-subgrid space-y-1">
			<div className="col-span-full flex w-full items-center justify-between">
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
						disabled={isEmptyArray(fields)}
						onClick={removeLast}
						size="icon-xs"
						variant="ghost"
					>
						<Minus />
					</Button>
				</ButtonGroup>
			</div>
			<div className="col-span-full grid grid-cols-subgrid gap-2">
				{fields.map(({ id }, index) => (
					<OpsGenieTarget
						defaults={usingDefaults ? defaults?.[index] : undefined}
						key={id}
						name={`${name}.${index}`}
						removeMe={
							fieldValues.length === 1
								? removeLast
								: () => {
										remove(index);
									}
						}
					/>
				))}
			</div>
		</div>
	);
};

export default memo(OpsGenieTargets);
