import { FC, memo } from 'react';
import { FormKeyValMap, FormText } from 'components/generic/form';
import { HeaderType, NotifyNtfyAction } from 'types/config';

interface Props {
	name: string;
	defaults?: NotifyNtfyAction;
}

/**
 * BROADCAST renders the form fields for the Broadcast Ntfy action.
 *
 * @param name - The name of the field in the form.
 * @param defaults - The default values for the Broadcast Ntfy action.
 * @returns The form fields for the Broadcast Ntfy action.
 */
const BROADCAST: FC<Props> = ({ name, defaults }) => (
	<>
		<FormText
			name={`${name}.intent`}
			col_sm={12}
			label="Intent"
			defaultVal={defaults?.intent}
			placeholder="e.g. 'io.heckel.ntfy.USER_ACTION'"
		/>
		<FormKeyValMap
			name={`${name}.extras`}
			label="Extras"
			tooltip="Android intent extras"
			defaults={defaults?.extras as HeaderType[] | undefined}
			keyPlaceholder="e.g. 'cmd'"
			valuePlaceholder="e.g. 'pic'"
		/>
	</>
);

export default memo(BROADCAST);
