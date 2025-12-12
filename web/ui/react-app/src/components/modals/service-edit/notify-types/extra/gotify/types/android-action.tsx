import type { FC } from 'react';
import { FieldText } from '@/components/generic/field.tsx';
import type { GotifyExtraProps } from '@/components/modals/service-edit/notify-types/extra/gotify/types/types.ts';
import { GOTIFY_EXTRA_NAMESPACE } from '@/utils/api/types/config/notify/gotify.ts';

const namespace = GOTIFY_EXTRA_NAMESPACE.ANDROID_ACTION.value;

/**
 * 'Android Action' extra type.
 *
 * @param name - The path to this extra in the form.
 * @param defaults - The defaults for this extra.
 * @returns The form fields for the 'Android Action' extra type at the given path in the form..
 */
export const AndroidAction: FC<GotifyExtraProps> = ({ name, defaults }) => {
	return (
		<FieldText
			colSize={{ md: 7, sm: 11, xs: 11 }}
			defaultVal={
				defaults?.namespace === namespace
					? defaults?.onReceive?.intentUrl
					: null
			}
			label="Intent URL"
			name={`${name}.onReceive.intentUrl`}
			placeholder="e.g. https://example.com"
			required
			tooltip={{
				content: 'Opens an intent after the notification was delivered',
				type: 'string',
			}}
		/>
	);
};
