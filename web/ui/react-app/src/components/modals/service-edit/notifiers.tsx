import { type FC, useCallback, useEffect, useMemo } from 'react';
import { useFieldArray, useFormContext } from 'react-hook-form';
import { createOption } from '@/components/generic/field-select-shared';
import Notify from '@/components/modals/service-edit/notify';
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
import { NOTIFY_TYPE_MAP } from '@/utils/api/types/config/notify/all-types';

type EditServiceNotifiersProps = {
	/* Whether the modal is loading. */
	loading: boolean;
};

/**
 * The form fields for a mutable list of notifiers.
 *
 * @param loading - Whether the modal is loading.
 * @returns The form fields for a mutable list of notifiers.
 */
const EditServiceNotifiers: FC<EditServiceNotifiersProps> = ({ loading }) => {
	const id = 'notify';
	const { mainDataDefaults, schemaDataDefaults } = useSchemaContext();
	const { setValue } = useFormContext();
	const { fields, append, remove } = useFieldArray({ name: id });

	const defaultArray = schemaDataDefaults?.notify;
	const mains = mainDataDefaults?.notify;

	// 'mains' that may be referenced.
	const globalNotifyOptions = useMemo(
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
				name: '',
				options: {},
				params: { avatar: '', color: '', icon: '' },
				type: Object.values(NOTIFY_TYPE_MAP)[0].value,
				url_fields: {},
			},
			{ shouldFocus: false },
		);
	}, []);

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
			<AccordionTrigger id={id}>Notify:</AccordionTrigger>
			<AccordionContent className={cn(!isEmptyArray(fields) && 'space-y-4')}>
				<Accordion className="w-full space-y-2" type="multiple">
					{fields.map(({ id: _id }, index) => (
						<AccordionItem key={_id} value={_id}>
							<Notify
								globalOptions={globalNotifyOptions}
								key={_id}
								mains={mains}
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
					Add Notify
				</Button>
			</AccordionContent>
		</AccordionItem>
	);
};

export default EditServiceNotifiers;
