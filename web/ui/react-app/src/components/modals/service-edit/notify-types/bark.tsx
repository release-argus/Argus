import { useEffect, useMemo } from 'react';
import { useFormContext } from 'react-hook-form';
import {
	FieldSelect,
	FieldText,
	FieldTextWithPreview,
} from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { normaliseForSelect } from '@/components/modals/service-edit/util';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import {
	type BarkScheme,
	type BarkSound,
	barkSchemeOptions,
	barkSoundOptions,
} from '@/utils/api/types/config/notify/bark';
import type { NotifyBarkSchema } from '@/utils/api/types/config-edit/notify/schemas';
import { nullString } from '@/utils/api/types/config-edit/shared/null-string';
import { ensureValue } from '@/utils/form-utils';

/**
 * The form fields for a `Bark` notifier.
 *
 * @param name - The path to this `Bark` in the form.
 * @param main - The main values.
 */
const BARK = ({ name, main }: { name: string; main?: NotifyBarkSchema }) => {
	const { getValues, setValue } = useFormContext();
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.bark,
		[main, typeDataDefaults?.notify.bark],
	);

	// Ensure selects have a valid value.
	// biome-ignore lint/correctness/useExhaustiveDependencies: fallback on first load.
	useEffect(() => {
		ensureValue<BarkScheme>({
			defaultValue: defaults?.params?.scheme,
			fallback: Object.values(barkSchemeOptions)[0].value,
			getValues,
			path: `${name}.params.scheme`,
			setValue,
		});
		ensureValue<BarkSound>({
			defaultValue: defaults?.params?.sound,
			fallback: Object.values(barkSoundOptions)[0].value,
			getValues,
			path: `${name}.params.sound`,
			setValue,
		});
	}, [main]);

	const barkSchemeOptionsNormalised = useMemo(() => {
		const defaultScheme = normaliseForSelect(
			barkSchemeOptions,
			defaults?.params?.scheme,
		);

		if (defaultScheme)
			return [
				{ label: `${defaultScheme.label} (default)`, value: nullString },
				...barkSchemeOptions,
			];

		return barkSchemeOptions;
	}, [defaults?.params?.scheme]);

	const barkSoundOptionsNormalised = useMemo(() => {
		const defaultSound = normaliseForSelect(
			barkSoundOptions,
			defaults?.params?.sound,
		);

		if (defaultSound)
			return [
				{ label: `${defaultSound.label} (default)`, value: nullString },
				...barkSoundOptions.filter((option) => option.value !== ''),
			];

		return barkSoundOptions;
	}, [defaults?.params?.sound]);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ xs: 9 }}
					defaultVal={defaults?.url_fields?.host}
					label="Host"
					name={`${name}.url_fields.host`}
					required
				/>
				<FieldText
					colSize={{ xs: 3 }}
					defaultVal={defaults?.url_fields?.port}
					label="Port"
					name={`${name}.url_fields.port`}
					required
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.path}
					label="Path"
					name={`${name}.url_fields.path`}
					tooltip={{
						content: 'Server path',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.devicekey}
					label="Device Key"
					name={`${name}.url_fields.devicekey`}
					required
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldSelect
					colSize={{ sm: 3 }}
					label="Scheme"
					name={`${name}.params.scheme`}
					options={barkSchemeOptionsNormalised}
					tooltip={{
						content: 'Server protocol',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 3 }}
					defaultVal={defaults?.params?.badge}
					label="Badge"
					name={`${name}.params.badge`}
					tooltip={{
						content: 'The number displayed next to the App icon',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.copy}
					label="Copy"
					name={`${name}.params.copy`}
					tooltip={{
						content: 'The value to be copied',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.group}
					label="Group"
					name={`${name}.params.group`}
					tooltip={{
						content: 'The group of the notification',
						type: 'string',
					}}
				/>
				<FieldSelect
					label="Sound"
					name={`${name}.params.sound`}
					options={barkSoundOptionsNormalised}
					required={false}
				/>
				<FieldText
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
				<FieldText
					defaultVal={defaults?.params?.url}
					label="URL"
					name={`${name}.params.url`}
					tooltip={{
						content: 'URL to open when notification is tapped',
						type: 'string',
					}}
				/>
				<FieldTextWithPreview
					defaultVal={defaults?.params?.icon}
					label="Icon"
					name={`${name}.params.icon`}
					tooltip={{
						content: 'URL to an icon',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default BARK;
