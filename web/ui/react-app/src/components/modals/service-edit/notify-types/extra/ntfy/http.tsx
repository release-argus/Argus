import { FC, memo } from 'react';
import {
	FormKeyValMap,
	FormSelect,
	FormText,
	FormTextArea,
} from 'components/generic/form';
import { HeaderType, NotifyNtfyAction } from 'types/config';

interface Props {
	name: string;
	defaults?: NotifyNtfyAction;
}

/**
 * HTTP renders the form fields for the HTTP Ntfy Action.
 *
 * @param name - The name of the field in the form.
 * @param defaults - The default values for the HTTP Ntfy Action.
 * @returns The form fields for this HTTP Ntfy Action.
 */
const HTTP: FC<Props> = ({ name, defaults }) => {
	const methodOptions = [
		{ label: 'POST', value: 'post' },
		{ label: 'PUT', value: 'put' },
		{ label: 'PATCH', value: 'patch' },
		{ label: 'GET', value: 'get' },
		{ label: 'DELETE', value: 'delete' },
	];

	return (
		<>
			<FormSelect
				name={`${name}.method`}
				col_sm={12}
				label="Type"
				options={methodOptions}
				positionXS="right"
			/>
			<FormText
				name={`${name}.url`}
				required
				col_sm={12}
				label="URL"
				defaultVal={defaults?.url}
				placeholder="e.g. 'https://ntfy.sh/mytopic'"
			/>
			<FormKeyValMap
				name={`${name}.headers`}
				label="Headers"
				tooltip="HTTP headers"
				defaults={defaults?.headers as HeaderType[] | undefined}
				keyPlaceholder="e.g. 'Authorization'"
				valuePlaceholder="e.g. 'Bearer <token>'"
			/>
			<FormTextArea
				name={`${name}.body`}
				col_sm={12}
				label="Body"
				defaultVal={defaults?.body}
				placeholder={`e.g. '{"key": "value"}'`}
			/>
		</>
	);
};

export default memo(HTTP);
