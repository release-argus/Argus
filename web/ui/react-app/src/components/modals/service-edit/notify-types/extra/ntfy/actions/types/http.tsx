import { type FC, memo } from 'react';
import {
	FieldCheck,
	FieldKeyValMap,
	FieldSelect,
	FieldText,
	FieldTextArea,
} from '@/components/generic/field';
import {
	NTFY_ACTION_TYPE,
	ntfyActionHTTPMethodOptions,
} from '@/utils/api/types/config/notify/ntfy';
import type { NtfyActionSchema } from '@/utils/api/types/config-edit/notify/types/ntfy';

type HTTPProps = {
	name: string;
	defaults?: NtfyActionSchema;
};

/**
 * Renders the form fields for the HTTP Ntfy Action.
 *
 * @param name - The name of the field in the form.
 * @param defaults - The default values for the HTTP Ntfy Action.
 * @returns The form fields for this HTTP Ntfy Action.
 */
const HTTP: FC<HTTPProps> = ({ name, defaults }) => {
	const itemDefaults =
		defaults?.action === NTFY_ACTION_TYPE.HTTP.value ? defaults : null;

	return (
		<>
			<FieldSelect
				colSize={{ md: 5, sm: 11 }}
				label="Method"
				name={`${name}.method`}
				options={ntfyActionHTTPMethodOptions}
				required
			/>
			<FieldText
				colSize={{ md: 10, sm: 9, xs: 9 }}
				defaultVal={itemDefaults?.url}
				label="URL"
				name={`${name}.url`}
				placeholder="e.g. 'https://ntfy.sh/mytopic'"
				required
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
			<FieldKeyValMap
				colSpan={11}
				defaults={itemDefaults?.headers}
				label="Headers"
				name={`${name}.headers`}
				placeholders={{
					key: "e.g. 'Authorization'",
					value: "e.g. 'Bearer <token>'",
				}}
				tooltip={{
					content: 'HTTP headers',
					type: 'string',
				}}
			/>
			<FieldTextArea
				colSize={{ sm: 12 }}
				defaultVal={itemDefaults?.body}
				label="Body"
				name={`${name}.body`}
				placeholder={`e.g. '{"key": "value"}'`}
			/>
		</>
	);
};

export default memo(HTTP);
