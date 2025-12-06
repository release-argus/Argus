import { Trash2 } from 'lucide-react';
import { type FC, useEffect, useMemo } from 'react';
import { useFormContext, useWatch } from 'react-hook-form';
import { BooleanWithDefault } from '@/components/generic';
import {
	FieldKeyValMap,
	FieldSelect,
	FieldText,
} from '@/components/generic/field';
import { AccordionContent, AccordionTrigger } from '@/components/ui/accordion';
import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import type { OptionType } from '@/components/ui/react-select/custom-components';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import { isEmptyOrNull } from '@/utils';
import {
	isWebHookType,
	type WebHookType,
	webhookTypeOptions,
} from '@/utils/api/types/config/webhook';

type EditServiceWebHookProps = {
	/* The name of the field in the form. */
	name: string;
	/* The function to remove this WebHook. */
	removeMe: () => void;

	/* The 'main' WebHook options. */
	globalOptions: OptionType[];
};

/**
 * The form fields for a WebHook.
 *
 * @param name - The name of the field in the form.
 * @param removeMe - The function to remove this WebHook.
 * @param globalOptions - The options for the global WebHooks.
 * @returns The form fields for this WebHook.
 */
const EditServiceWebHook: FC<EditServiceWebHookProps> = ({
	name,
	removeMe,

	globalOptions,
}) => {
	const { clearErrors, setValue, trigger } = useFormContext();
	const { mainDataDefaults, typeDataDefaults } = useSchemaContext();

	const itemName = useWatch({ name: `${name}.name` }) as string;
	// Main values.
	const main = mainDataDefaults?.webhook[itemName];
	// Default values.
	const itemType = useWatch({ name: `${name}.type` }) as WebHookType;
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.webhook,
		[main, typeDataDefaults],
	);

	// Sync type with main.
	// biome-ignore lint/correctness/useExhaustiveDependencies: clearErrors stable.
	useEffect(() => {
		// Set Type to that of the global for the new name (if one exists).
		let typeChanged = false;
		if (main?.type) {
			setValue(`${name}.type`, main.type);
			typeChanged = true;
		} else if (isWebHookType(itemName)) {
			setValue(`${name}.type`, itemName);
			typeChanged = true;
		}

		// Trigger validation on `name/type`.
		let timeout: ReturnType<typeof setTimeout> | undefined;
		if (typeChanged) {
			timeout = setTimeout(() => {
				clearErrors(name);
				if (!isEmptyOrNull(itemName)) void trigger(`${name}.name`);
				void trigger(`${name}.type`);
			}, 25);
		}

		return () => {
			if (timeout) clearTimeout(timeout);
		};
	}, [itemName, main]);

	// Accordion header.
	const header = useMemo(
		() => `${name.split('.').pop() ?? '-'}: (${itemType}) ${itemName}`,
		[name, itemName, itemType],
	);

	return (
		<>
			<ButtonGroup className="w-full">
				<Button
					aria-label="Delete this WebHook"
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
						content: 'Use this WebHook as a base (`webhook.x` in config root)',
						type: 'string',
					}}
				/>
				<FieldSelect
					colSize={{ xs: 6 }}
					label="Type"
					name={`${name}.type`}
					options={webhookTypeOptions}
					tooltip={{
						content: 'Style of WebHook to emulate',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					label={'Name'}
					name={`${name}.name`}
					required
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url}
					label="Target URL"
					name={`${name}.url`}
					required
					tooltip={{
						content: 'Where to send the WebHook',
						type: 'string',
					}}
					type="text"
				/>
				<BooleanWithDefault
					defaultValue={defaults?.allow_invalid_certs}
					label="Allow Invalid Certs"
					name={`${name}.allow_invalid_certs`}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.secret}
					label="Secret"
					name={`${name}.secret`}
					required
				/>
				<FieldKeyValMap
					defaults={defaults?.custom_headers}
					name={`${name}.custom_headers`}
				/>
				<FieldText
					colSize={{ xs: 6 }}
					defaultVal={defaults?.desired_status_code}
					label="Desired Status Code"
					name={`${name}.desired_status_code`}
					tooltip={{
						content:
							'Treat the WebHook as successful when this status code is received (0=2XX)',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ xs: 6 }}
					defaultVal={defaults?.max_tries}
					label="Max tries"
					name={`${name}.max_tries`}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.delay}
					label="Delay"
					name={`${name}.delay`}
					tooltip={{
						content: 'Delay sending by this duration',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.silent_fails}
					label="Silent fails"
					name={`${name}.silent_fails`}
					tooltip={{
						content: 'Notify if WebHook fails max tries times',
						type: 'string',
					}}
				/>
			</AccordionContent>
		</>
	);
};

export default EditServiceWebHook;
