import { Minus, Plus } from 'lucide-react';
import { type FC, memo, useCallback, useEffect, useMemo } from 'react';
import { useFieldArray, useFormContext, useWatch } from 'react-hook-form';
import { FieldLabel } from '@/components/generic/field';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import NtfyAction from '@/components/modals/service-edit/notify-types/extra/ntfy/actions/action';
import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import { isEmptyArray, isEmptyOrNull } from '@/utils';
import {
	NTFY_ACTION_TYPE,
	ntfyActionTypeOptions,
} from '@/utils/api/types/config/notify/ntfy';
import {
	type NtfyActionsSchema,
	ntfyActionsSchema,
} from '@/utils/api/types/config-edit/notify/types/ntfy';
import { isUsingDefaults } from '@/utils/api/types/config-edit/validators';

type BaseProps = {
	/* The name of the field in the form. */
	name: string;
	/* The label for the field. */
	label: string;
	/* The default values for the field. */
	defaults?: NtfyActionsSchema;
};

type NtfyActionsProps = BaseProps & {
	/* The tooltip on the field label. */
	tooltip: TooltipWithAriaProps;
};

/**
 * NtfyActions returns the form fields for the Ntfy actions.
 *
 * @param name - The name of the field in the form.
 * @param label - The label for the field.
 * @param tooltip - The tooltip on the field label.
 * @param tooltip.type - 'string' | 'element'.
 * @param tooltip.side - The wide to render the tooltip content.
 * @param tooltip.size - The size of the tooltip.
 * @param tooltip.delayDuration - Time before rendering the tooltip.
 * @param defaults - The default values for the field.
 * @returns A set of form fields for a list of Ntfy actions.
 */
const NtfyActions: FC<NtfyActionsProps> = ({
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
				action: Object.values(ntfyActionTypeOptions)[0].value,
			},
			{ shouldFocus: false },
		);
	}, []);

	const defaultsHollow = (defaults ?? []).map((a) => {
		switch (a.action) {
			case NTFY_ACTION_TYPE.HTTP.value:
				return {
					action: a.action,
					method: a.method,
				};
			case NTFY_ACTION_TYPE.VIEW.value:
				return {
					action: a.action,
					clear: a.clear,
				};
			default: {
				// Broadcast
				return { action: a.action };
			}
		}
	});

	// Keep track of the array values, so we can switch to defaults when unchanged.
	const fieldValues = useWatch({ name: name }) as NtfyActionsSchema | undefined;
	// usingDefaults when fieldValues unset or match the defaults.
	const usingDefaults = useMemo(
		() =>
			isUsingDefaults({
				arg: fieldValues,
				defaultValue: defaults,
				matchingFieldsEndsWiths: ['.action', '.method', '.clear'],
				schema: ntfyActionsSchema,
			}),
		[fieldValues, defaults],
	);
	// Validate on change of defaults being usable.
	// Give defaults back when empty.
	// biome-ignore lint/correctness/useExhaustiveDependencies: usingDefaults covers fieldValues.
	useEffect(() => {
		trigger(name).catch(() => {
			return;
		});

		// Give defaults back if field empty.
		if (isEmptyArray(fieldValues) && defaultsHollow) {
			setValue(name, defaultsHollow);
		}
	}, [usingDefaults]);

	// On load, ensure we don't have actions from a different type
	// and give the defaults if not overridden.
	// biome-ignore lint/correctness/useExhaustiveDependencies: defaultsHollow stable.
	useEffect(() => {
		for (const item of fieldValues ?? []) {
			if (isEmptyOrNull(item.action)) {
				setValue(name, defaultsHollow);
				break;
			}
		}
	}, []);

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
					<NtfyAction
						defaults={usingDefaults ? defaults?.[index] : undefined}
						key={id}
						name={`${name}.${index}`}
						removeMe={
							// Give a disabled remove if only one item, and it matches the defaults.
							fieldValues?.length === 1
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

export default memo(NtfyActions);
