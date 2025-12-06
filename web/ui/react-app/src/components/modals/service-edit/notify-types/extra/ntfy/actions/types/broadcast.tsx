import { type FC, memo } from 'react';
import {
	FieldCheck,
	FieldKeyValMap,
	FieldText,
} from '@/components/generic/field';
import { NTFY_ACTION_TYPE } from '@/utils/api/types/config/notify/ntfy';
import type { NtfyActionSchema } from '@/utils/api/types/config-edit/notify/types/ntfy';

type BROADCASTProps = {
	name: string;
	defaults?: NtfyActionSchema;
};

/**
 * Renders the form fields for the BROADCAST Ntfy action.
 *
 * @param name - The name of the field in the form.
 * @param defaults - The default values for the Broadcast Ntfy action.
 * @returns The form fields for the Broadcast Ntfy action.
 */
const BROADCAST: FC<BROADCASTProps> = ({ name, defaults }) => {
	const itemDefaults =
		defaults?.action === NTFY_ACTION_TYPE.BROADCAST.value ? defaults : null;

	return (
		<>
			<FieldText
				colSize={{ md: 4, sm: 9, xs: 9 }}
				defaultVal={itemDefaults?.intent}
				label="Intent"
				name={`${name}.intent`}
				placeholder="e.g. 'io.heckel.ntfy.USER_ACTION'"
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
				defaults={itemDefaults?.extras}
				label="Extras"
				name={`${name}.extras`}
				placeholders={{ key: "e.g. 'cmd'", value: "e.g. 'pic'" }}
				tooltip={{
					content: 'Android intent extras',
					type: 'string',
				}}
			/>
		</>
	);
};

export default memo(BROADCAST);
