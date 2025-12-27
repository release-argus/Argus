import type { FC } from 'react';
import { FieldCheck, FieldText } from '@/components/generic/field';
import { NTFY_ACTION_TYPE } from '@/utils/api/types/config/notify/ntfy';
import type { NtfyActionSchema } from '@/utils/api/types/config-edit/notify/types/ntfy';

type VIEWProps = {
	name: string;
	defaults?: NtfyActionSchema;
};

/**
 * Renders the form fields for the VIEW Ntfy action.
 *
 * @param name - The name of the field in the form.
 * @param defaults - The default values for the action.
 * @returns The form fields for this action.
 */
const VIEW: FC<VIEWProps> = ({ name, defaults }) => {
	const itemDefaults =
		defaults?.action === NTFY_ACTION_TYPE.VIEW.value ? defaults : null;

	return (
		<>
			<FieldText
				colSize={{ md: 4, sm: 9, xs: 9 }}
				defaultVal={itemDefaults?.url}
				label="URL"
				name={`${name}.url`}
				placeholder="e.g. 'https://example.com'"
				required
				tooltip={{
					content: 'URL to open when action is pressed',
					type: 'string',
				}}
			/>
			<FieldCheck
				colSize={{ md: 1, sm: 2, xs: 2 }}
				label="Clear"
				name={`${name}.clear`}
				tooltip={{
					content: 'Clear notification after action is pressed',
					type: 'string',
				}}
			/>
		</>
	);
};

export default VIEW;
