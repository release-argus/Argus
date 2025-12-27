import { Trash2 } from 'lucide-react';
import type { FC } from 'react';
import { useWatch } from 'react-hook-form';
import { FieldSelect } from '@/components/generic/field';
import RenderURLCommand from '@/components/modals/service-edit/url-commands/render';
import { Button } from '@/components/ui/button';
import {
	latestVersionURLCommandTypeOptions,
	type URLCommand,
} from '@/utils/api/types/config/service/latest-version';

type FormURLCommandProps = {
	name: string;
	removeMe: () => void;
};

/**
 * The form fields for a URL command.
 *
 * @param name - The name of the field in the form.
 * @param removeMe - The function to remove the command.
 * @returns The form fields for this URL command.
 */
const FormURLCommand: FC<FormURLCommandProps> = ({ name, removeMe }) => {
	const commandType = useWatch({ name: `${name}.type` }) as URLCommand['type'];

	return (
		<div className="col-span-full grid grid-cols-subgrid gap-x-2">
			<div className="col-span-1 py-1 md:pe-2">
				<Button
					aria-label="Delete this URL command"
					className="size-full"
					onClick={removeMe}
					variant="outline"
				>
					<Trash2 />
				</Button>
			</div>
			<div className="col-span-11 grid grid-cols-subgrid gap-2">
				<FieldSelect
					colSize={{ sm: 3, xs: 4 }}
					label="Type"
					labelSize="sm"
					name={`${name}.type`}
					options={latestVersionURLCommandTypeOptions}
				/>
				{<RenderURLCommand commandType={commandType} name={name} />}
			</div>
		</div>
	);
};

export default FormURLCommand;
