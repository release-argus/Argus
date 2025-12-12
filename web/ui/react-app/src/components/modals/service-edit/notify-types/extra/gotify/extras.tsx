import { Minus, Plus } from 'lucide-react';
import { type FC, memo, useCallback, useEffect, useMemo } from 'react';
import { useFieldArray, useFormContext, useWatch } from 'react-hook-form';
import { FieldLabel } from '@/components/generic/field';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import Extra from '@/components/modals/service-edit/notify-types/extra/gotify/extra';
import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import { isEmptyArray } from '@/utils';
import {
	GOTIFY_EXTRA_NAMESPACE,
	gotifyExtraNamespaceOptions,
} from '@/utils/api/types/config/notify/gotify.ts';
import {
	type GotifyExtrasSchema,
	gotifyExtrasSchema,
} from '@/utils/api/types/config-edit/notify/types/gotify.ts';
import { isUsingDefaults } from '@/utils/api/types/config-edit/validators.ts';

type BaseProps = {
	/* The name of the field in the form. */
	name: string;
	/* The label for the field. */
	label?: string;

	/* The default values for the field. */
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	defaults?: GotifyExtrasSchema;
};

type ExtrasProps = BaseProps & {
	/* Optional tooltip for the label. */
	tooltip?: TooltipWithAriaProps;
};

/**
 * Extras renders the form fields for a list of Gotify extras.
 *
 * @param name - The name of the field in the form.
 * @param label - The label for the field.
 * @param tooltip - The tooltip on the field label.
 * @param tooltip.type - 'string' | 'element'.
 * @param tooltip.side - The wide to render the tooltip content.
 * @param tooltip.size - The size of the tooltip.
 * @param tooltip.delayDuration - Time before rendering the tooltip.
 * @param defaults - The default values for the field.
 * @returns A set of form fields for a list of Extras.
 */
const Extras: FC<ExtrasProps> = ({
	name,
	label = 'Extras',
	tooltip,
	defaults,
}) => {
	const { setValue, trigger } = useFormContext();
	const { fields, append, remove } = useFieldArray({ name: name });

	// biome-ignore lint/correctness/useExhaustiveDependencies: append stable.
	const addItem = useCallback(() => {
		append(
			{
				bigImageUrl: '',
				namespace: Object.values(gotifyExtraNamespaceOptions)[0].value,
			},
			{ shouldFocus: false },
		);
	}, []);

	const fieldValues = useWatch({ name }) as GotifyExtrasSchema;
	const usingDefaults = useMemo(
		() =>
			isUsingDefaults({
				arg: fieldValues,
				defaultValue: defaults,
				matchingFieldsEndsWiths: ['.namespace', '.contentType'],
				schema: gotifyExtrasSchema,
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
		if (defaults && isEmptyArray(fieldValues)) {
			// Need to explicitly remove (empty) all fields to prevent RHF using defaultValues.
			const hollowDefaults = gotifyExtrasSchema.parse(
				defaults.map((e) =>
					e.namespace === GOTIFY_EXTRA_NAMESPACE.CLIENT_DISPLAY.value
						? { contentType: e.contentType, namespace: e.namespace }
						: { namespace: e.namespace },
				),
			);
			setValue(name, hollowDefaults);
		}
	}, [usingDefaults]);

	// Remove the last item if not the only one, or doesn't match the defaults.
	// biome-ignore lint/correctness/useExhaustiveDependencies: remove stable.
	const removeLast = useCallback(() => {
		if (!(usingDefaults && fields.length === 1)) remove(fields.length - 1);
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
					<Extra
						defaults={usingDefaults ? defaults?.[index] : undefined}
						key={id}
						name={`${name}.${index}`}
						removeMe={
							fieldValues?.length === 1 ? removeLast : () => remove(index)
						}
					/>
				))}
			</div>
		</div>
	);
};

export default memo(Extras);
