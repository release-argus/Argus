import { type FC, useEffect } from 'react';
import { useFormContext } from 'react-hook-form';
import { FieldSelect } from '@/components/generic/field.tsx';
import type { GotifyExtraProps } from '@/components/modals/service-edit/notify-types/extra/gotify/types/types.ts';
import {
	GOTIFY_EXTRA_NAMESPACE,
	type GotifyExtraClientDisplayContentType,
	gotifyExtraClientDisplayContentTypeOptions,
} from '@/utils/api/types/config/notify/gotify.ts';
import { ensureValue } from '@/utils/form-utils.ts';

const namespace = GOTIFY_EXTRA_NAMESPACE.CLIENT_DISPLAY.value;

/**
 * 'Client display' extra type.
 *
 * @param name - The path to this extra in the form.
 * @param defaults - The defaults for this extra.
 * @returns The form fields for the 'Client display' extra type at the given path in the form..
 */
export const ClientDisplay: FC<GotifyExtraProps> = ({ name, defaults }) => {
	const { getValues, setValue } = useFormContext();

	// Ensure selects have a valid value.
	// biome-ignore lint/correctness/useExhaustiveDependencies: fallback on first load.
	useEffect(() => {
		ensureValue<GotifyExtraClientDisplayContentType>({
			defaultValue:
				defaults?.namespace === namespace ? defaults?.contentType : null,
			fallback: Object.values(gotifyExtraClientDisplayContentTypeOptions)[0]
				.value,
			getValues,
			path: `${name}.contentType`,
			setValue,
		});
	}, [defaults]);

	return (
		<FieldSelect
			colSize={{ md: 7, xs: 6 }}
			label="Content Type"
			name={`${name}.contentType`}
			options={gotifyExtraClientDisplayContentTypeOptions}
			required
		/>
	);
};
