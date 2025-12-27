import { type FC, useCallback, useEffect, useMemo } from 'react';
import { useFieldArray, useFormContext, useWatch } from 'react-hook-form';
import Command from '@/components/modals/service-edit/command';
import {
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from '@/components/ui/accordion';
import { Button } from '@/components/ui/button';
import { FieldSet } from '@/components/ui/field';
import { Separator } from '@/components/ui/separator';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import { cn } from '@/lib/utils';
import { isEmptyArray } from '@/utils';
import type { CommandsSchema } from '@/utils/api/types/config-edit/command/schemas';
import { isUsingDefaults } from '@/utils/api/types/config-edit/validators';

type EditServiceCommandsProps = {
	/* The name of the field in the form. */
	name: string;
	/* Whether the modal is loading. */
	loading: boolean;
};

/**
 * The form fields for all commands in a service.
 *
 * @param name - The name of the field in the form.
 * @param loading - Whether the modal is loading.
 * @returns The set of form fields for a list of `command`.
 */
const EditServiceCommands: FC<EditServiceCommandsProps> = ({
	name,
	loading,
}) => {
	const { schemaDataDefaults, schemaDataDefaultsHollow } = useSchemaContext();
	const defaults = schemaDataDefaults?.command;
	const defaultsHollow = schemaDataDefaultsHollow?.command;
	const { setValue, trigger } = useFormContext();
	const { fields, append, remove } = useFieldArray({ name: name });

	// biome-ignore lint/correctness/useExhaustiveDependencies: append stable.
	const addItem = useCallback(() => {
		append([[{ arg: '' }]], { shouldFocus: false });
	}, []);

	// Remove item at given index.
	// biome-ignore lint/correctness/useExhaustiveDependencies: remove stable.
	const removeItem = useCallback(
		(index: number) => () => {
			// Change focus to the accordion.
			document.getElementById(name)?.focus();
			remove(index);
		},
		[],
	);

	// Keep track of the array values, so we can use defaults when empty.
	const fieldValues = useWatch({ name: name }) as CommandsSchema;
	// Use defaults when fieldValues undefined or the same as the defaults.
	const usingDefaults = useMemo(
		() => isUsingDefaults({ arg: fieldValues, defaultValue: defaults }),
		[defaults, fieldValues],
	);

	// Trigger validation on change of defaults used/not.
	// Reset to defaults when empty.
	// biome-ignore lint/correctness/useExhaustiveDependencies: usingDefaults covers fieldValues.
	useEffect(() => {
		void trigger(name);

		// Give defaults back when field empty.
		if (!isEmptyArray(defaultsHollow) && isEmptyArray(fieldValues)) {
			setValue(name, defaultsHollow);
		}
	}, [usingDefaults, defaultsHollow]);

	return (
		<AccordionItem value={name}>
			<AccordionTrigger id={name}>Command:</AccordionTrigger>
			<AccordionContent className="mb-2 flex flex-col gap-2">
				{fields.map(({ id }, index) => (
					<FieldSet className="grid grid-cols-12 gap-2" key={id}>
						<Command
							defaults={usingDefaults ? defaults?.[index] : undefined}
							name={`${name}.${index}`}
							removeMe={removeItem(index)}
						/>
						{index < fields.length - 1 && (
							<Separator className="col-span-12 my-2" />
						)}
					</FieldSet>
				))}
				<Button
					className={cn(isEmptyArray(fields) && 'mt-2', 'w-full')}
					disabled={loading}
					onClick={addItem}
					variant="secondary"
				>
					Add Command
				</Button>
			</AccordionContent>
		</AccordionItem>
	);
};

export default EditServiceCommands;
