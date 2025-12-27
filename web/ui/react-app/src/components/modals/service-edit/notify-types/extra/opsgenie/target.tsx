import { Trash2 } from 'lucide-react';
import type { FC } from 'react';
import { useWatch } from 'react-hook-form';
import { FieldSelect, FieldText } from '@/components/generic/field';
import { Button } from '@/components/ui/button';
import {
	OpsGenieTargetSubTypeOptions,
	OpsGenieTargetTypeOptions,
} from '@/utils/api/types/config/notify/opsgenie';
import type { OpsGenieTargetSchema } from '@/utils/api/types/config-edit/notify/types/opsgenie';

type OpsGenieTargetProps = {
	/* The name of the field in the form. */
	name: string;
	/* The function to remove this target. */
	removeMe: () => void;

	/* The default values for the target. */
	defaults?: OpsGenieTargetSchema;
};

/**
 * OpsGenieTarget renders fields for an OpsGenie target.
 *
 * @param name - The name of the field in the form.
 * @param removeMe - The function to remove this target.
 * @param defaults - The default values for the target.
 * @returns The form fields for this OpsGenie target.
 */
const OpsGenieTarget: FC<OpsGenieTargetProps> = ({
	name,
	removeMe,
	defaults,
}) => {
	const targetType = useWatch({
		name: `${name}.type`,
	}) as OpsGenieTargetSchema['type'];

	return (
		<div className="col-span-full grid grid-cols-subgrid gap-x-2">
			<div className="col-span-1 py-1 md:pe-2">
				<Button
					aria-label="Remove this target"
					className="size-full"
					onClick={removeMe}
					size="icon-md"
					variant="outline"
				>
					<Trash2 />
				</Button>
			</div>
			<div className="col-span-11 grid grid-cols-subgrid gap-x-2">
				<FieldSelect
					colSize={{ md: 2, sm: 5, xs: 5 }}
					name={`${name}.type`}
					options={OpsGenieTargetTypeOptions}
				/>
				<FieldSelect
					colSize={{ md: 3, sm: 6, xs: 6 }}
					name={`${name}.sub_type`}
					options={OpsGenieTargetSubTypeOptions[targetType]}
				/>
				<FieldText
					colSize={{ md: 6, sm: 11, xs: 11 }}
					defaultVal={defaults?.value}
					name={`${name}.value`}
					required
				/>
			</div>
		</div>
	);
};

export default OpsGenieTarget;
