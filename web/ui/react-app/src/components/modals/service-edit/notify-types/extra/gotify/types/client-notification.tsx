import type { FC } from 'react';
import {
	FieldText,
	FieldTextWithPreview,
} from '@/components/generic/field.tsx';
import type { GotifyExtraProps } from '@/components/modals/service-edit/notify-types/extra/gotify/types/types.ts';
import { GOTIFY_EXTRA_NAMESPACE } from '@/utils/api/types/config/notify/gotify.ts';

const namespace = GOTIFY_EXTRA_NAMESPACE.CLIENT_NOTIFICATION.value;

/**
 * 'Client notification' extra type.
 *
 * @param name - The path to this extra in the form.
 * @param defaults - The defaults for this extra.
 * @returns The form fields for the 'Client notification' extra type at the given path in the form..
 */
export const ClientNotification: FC<GotifyExtraProps> = ({
	name,
	defaults,
}) => {
	return (
		<>
			<FieldText
				colSize={{ md: 7, sm: 11, xs: 11 }}
				defaultVal={
					defaults?.namespace === namespace ? defaults?.click?.url : null
				}
				label="Click URL"
				name={`${name}.click.url`}
				placeholder="e.g. https://example.com"
				tooltip={{
					content: 'Opens a URL on notification click',
					type: 'string',
				}}
			/>
			<FieldTextWithPreview
				colSize={{ md: 11, sm: 11, xs: 11 }}
				defaultVal={
					defaults?.namespace === namespace ? defaults?.bigImageUrl : undefined
				}
				label="Big Image URL"
				name={`${name}.bigImageUrl`}
				placeholder="e.g. https://example.com/icon.png"
				tooltip={{
					content: 'Shows a big image in the notification.',
					type: 'string',
				}}
			/>
		</>
	);
};
