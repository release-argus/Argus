import { FC, memo } from 'react';

import { FormText } from 'components/generic/form';
import { NotifyNtfyAction } from 'types/config';

interface Props {
	name: string;
	defaults?: NotifyNtfyAction;
}

/**
 * VIEW renders the form fields for the Ntfy action.
 *
 * @param name - The name of the field in the form.
 * @param defaults - The default values for the action.
 * @returns The form fields for this action.
 */
const VIEW: FC<Props> = ({ name, defaults }) => (
	<FormText
		name={`${name}.url`}
		required
		col_sm={12}
		label="URL"
		defaultVal={defaults?.url}
		placeholder="e.g. 'https://example.com'"
	/>
);

export default memo(VIEW);
