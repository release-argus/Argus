import { Trash2 } from 'lucide-react';
import type { FC } from 'react';
import { useWatch } from 'react-hook-form';
import { FieldSelect, FieldText } from '@/components/generic/field';
import RenderAction from '@/components/modals/service-edit/notify-types/extra/ntfy/actions/render';
import { Button } from '@/components/ui/button';
import { ntfyActionTypeOptions } from '@/utils/api/types/config/notify/ntfy';
import type { NtfyActionSchema } from '@/utils/api/types/config-edit/notify/types/ntfy';

type NtfyActionProps = {
	/* The name of the field in the form. */
	name: string;
	/* The default values for the action. */
	defaults?: NtfyActionSchema;
	/* The function to remove this action. */
	removeMe: () => void;
};

/**
 * The form fields for a 'NTFY' action.
 *
 * @param name - The name of the field in the form.
 * @param defaults - The default values for the action.
 * @param removeMe - The function to remove this action.
 */
const NtfyAction: FC<NtfyActionProps> = ({ name, defaults, removeMe }) => {
	const typeLabelMap = {
		broadcast: 'Take picture',
		http: 'Close door',
		view: 'Open page',
	};

	const targetType = useWatch({
		name: `${name}.action`,
	}) as keyof typeof typeLabelMap;

	return (
		<div className="col-span-full grid grid-cols-subgrid gap-x-2">
			<div className="col-span-1 py-1 md:pe-2">
				<Button
					aria-label="Remove this action"
					className="size-full"
					onClick={removeMe}
					size="icon-md"
					variant="outline"
				>
					<Trash2 />
				</Button>
			</div>
			<div className="col-span-11 grid grid-cols-subgrid gap-x-2 gap-y-1">
				<FieldSelect
					colSize={{ md: 3, sm: 5, xs: 5 }}
					label="Action Type"
					name={`${name}.action`}
					options={ntfyActionTypeOptions}
					required
				/>
				<FieldText
					colSize={{ md: 3, sm: 6, xs: 6 }}
					defaultVal={defaults?.label}
					label="Label"
					name={`${name}.label`}
					placeholder={`e.g. '${typeLabelMap[targetType]}'`}
					required
					tooltip={{
						content: 'Button name to display on the notification',
						type: 'string',
					}}
				/>
				<RenderAction defaults={defaults} name={name} targetType={targetType} />
			</div>
		</div>
	);
};

export default NtfyAction;
