import type { FC } from 'react';
import { FieldText } from '@/components/generic/field.tsx';
import type { GotifyExtraProps } from '@/components/modals/service-edit/notify-types/extra/gotify/types/types.ts';
import { GOTIFY_EXTRA_NAMESPACE } from '@/utils/api/types/config/notify/gotify.ts';

const namespace = GOTIFY_EXTRA_NAMESPACE.OTHER.value;

/**
 * Other (custom) extra type.
 *
 * @param name - The path to this extra in the form.
 * @param defaults - The defaults for this extra.
 * @returns The form fields for the 'Other' extra type at the given path in the form..
 */
export const Other: FC<GotifyExtraProps> = ({ name, defaults }) => {
	return (
		<>
			<FieldText
				colSize={{ md: 7, sm: 6, xs: 6 }}
				defaultVal={
					defaults?.namespace === namespace ? defaults?._namespace : null
				}
				label="Name"
				name={`${name}._namespace`}
				required
				tooltip={{
					content: 'Namespace for this extra',
					type: 'string',
				}}
			/>
			<FieldText
				colSize={{ md: 11, sm: 12, xs: 12 }}
				defaultVal={defaults?.namespace === namespace ? defaults?.value : null}
				label="Value"
				name={`${name}.value`}
				required
				tooltip={{
					content: 'Full JSON value for this extra (under the namespace)',
					type: 'string',
				}}
			/>
		</>
	);
};
