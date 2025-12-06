import { type FC, useCallback, useEffect, useMemo } from 'react';
import { useFieldArray, useFormContext } from 'react-hook-form';
import { createOption } from '@/components/generic/field-select-shared';
import EditServiceWebHook from '@/components/modals/service-edit/webhook';
import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from '@/components/ui/accordion';
import { Button } from '@/components/ui/button';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import { cn } from '@/lib/utils';
import { isEmptyArray } from '@/utils';
import { WEBHOOK_TYPE } from '@/utils/api/types/config/webhook';

type EditServiceWebHooksProps = {
	/* Whether the modal is loading. */
	loading: boolean;
};

/**
 * The form fields for s service's webhooks.
 *
 * @param loading - Whether the modal is loading.
 * @returns The form fields for a service's webhooks.
 */
const EditServiceWebHooks: FC<EditServiceWebHooksProps> = ({ loading }) => {
	const id = 'webhook';
	const {
		mainDataDefaults,
		schemaDataDefaults,
		typeDataDefaults,
		typeDataDefaultsHollow,
	} = useSchemaContext();
	const { setValue } = useFormContext();
	const { fields, append, remove } = useFieldArray({ name: id });

	const defaultArray = schemaDataDefaults?.webhook;
	const mains = mainDataDefaults?.webhook;
	const defaults = typeDataDefaults?.webhook;
	const defaultsHollow = typeDataDefaultsHollow?.webhook;

	// 'mains' that may be referenced.
	const globalWebHookOptions = useMemo(
		() => [
			{ label: '--None--', value: '' },
			...Object.keys(mains ?? []).map((n) => createOption(n)),
		],
		[mains],
	);

	// biome-ignore lint/correctness/useExhaustiveDependencies: append stable.
	const addItem = useCallback(() => {
		append(
			{
				custom_headers: defaultsHollow?.custom_headers,
				name: '',
				type: Object.values(WEBHOOK_TYPE)[0].value,
			},
			{ shouldFocus: false },
		);
	}, [defaults]);

	// Remove item at given index.
	// biome-ignore lint/correctness/useExhaustiveDependencies: remove stable.
	const removeItem = useCallback(
		(index: number) => () => {
			// Change focus to the accordion.
			document.getElementById(id)?.focus();
			remove(index);
		},
		[],
	);

	// Reset to defaults when empty.
	// biome-ignore lint/correctness/useExhaustiveDependencies: setValue stable.
	useEffect(() => {
		if (fields.length === 0 && (defaultArray ?? [])?.length > 0)
			setValue(id, defaultArray);
	}, [defaultArray, fields.length]);

	return (
		<AccordionItem value={id}>
			<AccordionTrigger id={id}>WebHook:</AccordionTrigger>
			<AccordionContent className={cn(!isEmptyArray(fields) && 'space-y-4')}>
				<Accordion className="w-full space-y-2" type="multiple">
					{fields.map(({ id: _id }, index) => (
						<AccordionItem key={_id} value={_id}>
							<EditServiceWebHook
								globalOptions={globalWebHookOptions}
								key={_id}
								name={`${id}.${index.toString()}`}
								removeMe={removeItem(index)}
							/>
						</AccordionItem>
					))}
				</Accordion>
				<Button
					className="w-full"
					disabled={loading}
					onClick={addItem}
					variant="secondary"
				>
					Add WebHook
				</Button>
			</AccordionContent>
		</AccordionItem>
	);
};

export default EditServiceWebHooks;
