import { Trash2 } from 'lucide-react';
import { type FC, memo, useCallback, useEffect, useMemo } from 'react';
import { useFormContext, useWatch } from 'react-hook-form';
import { FieldSelect, FieldText } from '@/components/generic/field';
import RenderNotify from '@/components/modals/service-edit/notify-types/render';
import TestNotify from '@/components/modals/service-edit/test-notify';
import { AccordionContent, AccordionTrigger } from '@/components/ui/accordion';
import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import type { OptionType } from '@/components/ui/react-select/custom-components';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import { isEmptyOrNull } from '@/utils';
import { notifyTypeOptions } from '@/utils/api/types/config/notify/all-types';
import {
	LATEST_VERSION_LOOKUP_TYPE,
	type LatestVersionLookupType,
} from '@/utils/api/types/config/service/latest-version';
import {
	isNotifyType,
	type NotifySchemaKeys,
	type NotifySchemaValues,
} from '@/utils/api/types/config-edit/notify/schemas';

type NotifyProps = {
	/* The name of the field in the form. */
	name: string;
	/* The function to remove this Notify. */
	removeMe: () => void;

	/* The 'main' notifier options. */
	globalOptions: OptionType[];
	/* The 'main' notifiers available to reference. */
	mains?: Record<string, NotifySchemaValues>;
};

/**
 * The form fields for a notifier.
 *
 * @param name - The name of the field in the form.
 * @param removeMe - The function to remove this Notify.
 * @param originals - The original values for the Notify.
 * @param globalOptions - The options for the global Notifiers.
 * @param mains - The main Notifiers.
 * @returns The form fields for this Notify.
 */
const Notify: FC<NotifyProps> = ({
	name,
	removeMe,

	globalOptions,
	mains,
}) => {
	const { schemaData, typeDataDefaults } = useSchemaContext();
	const { clearErrors, setValue, trigger } = useFormContext();

	// Original values.
	const originals = schemaData?.notify;
	// Main values.
	const itemName = useWatch({ name: `${name}.name` }) as string;
	const main = useMemo(() => mains?.[itemName], [mains, itemName]);
	// Default values.
	const itemType = useWatch({ name: `${name}.type` }) as NotifySchemaKeys;
	const defaults = typeDataDefaults?.notify[itemType];

	const lvType = useWatch({
		name: 'latest_version.type',
	}) as LatestVersionLookupType;
	const lvURL = useWatch({ name: 'latest_version.url' }) as string | undefined;
	const webURL = useWatch({ name: 'dashboard.web_url' }) as string | undefined;

	// Sync type with main.
	// biome-ignore lint/correctness/useExhaustiveDependencies: main.type doesn't change without itemName.
	useEffect(() => {
		// Set Type to that of the global for the new name if it exists.
		if (main?.type) {
			setValue(`${name}.type`, main.type);
		} else if (isNotifyType(itemName)) {
			setValue(`${name}.type`, itemName);
		}

		// Trigger validation on `name/type`.
		const timeout = setTimeout(() => {
			if (!isEmptyOrNull(itemName)) void trigger(`${name}.name`);
			void trigger(`${name}.type`);
		}, 25);

		return () => {
			clearTimeout(timeout);
		};
	}, [itemName, mains]);

	// Clear errors and trigger validation when main/type changes.
	// biome-ignore lint/correctness/useExhaustiveDependencies: clearErrors stable.
	useEffect(() => {
		clearErrors(name);
		void trigger(`${name}.type`);
	}, [main, itemType]);

	// Original values for this notify element.
	const original = useMemo(
		() => originals?.find((o) => o.old_index === itemName),
		[itemName, originals],
	);
	const serviceURL =
		lvType === LATEST_VERSION_LOOKUP_TYPE.GITHUB.value &&
		(lvURL?.match(/\//g) ?? []).length == 1
			? `https://github.com/${lvURL ?? ''}`
			: lvURL;

	// Reset values to their original when type reverted.
	const onChangeNotifyType = useCallback(
		(newType: NotifySchemaKeys) => {
			// Reset to the original type.
			if (original && newType === original?.type) {
				setValue(`${name}.url_fields`, original.url_fields);
				setValue(`${name}.params`, original.params);
			}
		},
		[name, original, setValue],
	);

	// Accordion header.
	const header = useMemo(
		() => `${name.split('.').slice(-1).join('')}: (${itemType}) ${itemName}`,
		[name, itemName, itemType],
	);

	return (
		<>
			<ButtonGroup className="w-full">
				<Button
					aria-label="Delete this notifier"
					onClick={removeMe}
					variant="ghost"
				>
					<Trash2 />
				</Button>
				<AccordionTrigger className="w-full items-center rounded-l-none py-0 pl-2">
					{header}
				</AccordionTrigger>
			</ButtonGroup>

			<AccordionContent className="grid grid-cols-12 gap-2 p-4">
				<FieldSelect
					colSize={{ xs: 6 }}
					label="Base"
					name={`${name}.name`}
					options={globalOptions}
					showError={false}
					tooltip={{
						content: 'Use this Notify as a base (`notify.x` in config root)',
						type: 'string',
					}}
				/>
				<FieldSelect
					colSize={{ xs: 6 }}
					label="Type"
					name={`${name}.type`}
					onChange={(newValue) => {
						const newType = newValue?.value as NotifySchemaKeys;
						onChangeNotifyType(newType);
						return newValue;
					}}
					options={notifyTypeOptions}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					label="Name"
					name={`${name}.name`}
					required
				/>
				<RenderNotify
					defaults={defaults}
					main={mains?.[itemName]}
					name={name}
					type={itemType}
				/>
				<TestNotify
					extras={{
						service_url: serviceURL,
						web_url: webURL,
					}}
					original={original}
					path={name}
				/>
			</AccordionContent>
		</>
	);
};

export default memo(Notify);
