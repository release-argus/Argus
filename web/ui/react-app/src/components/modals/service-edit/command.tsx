import { Minus, Plus, Trash2 } from 'lucide-react';
import { type FC, useCallback } from 'react';
import { useFieldArray } from 'react-hook-form';
import { FieldText } from '@/components/generic/field';
import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import { FieldGroup } from '@/components/ui/field';
import { isEmptyArray } from '@/utils';
import type { CommandSchema } from '@/utils/api/types/config-edit/command/schemas';

type CommandProps = {
	/* The name of the field in the form. */
	name: string;
	/* The default values for the command. */
	defaults?: CommandSchema;
	/* The function to remove the command. */
	removeMe?: () => void;
};

/**
 * The form fields for a command.
 *
 * @param name - The name of the field in the form.
 * @param defaults - The default values for the command.
 * @param removeMe - The function to remove the command.
 * @returns The form fields for this command with any number of arguments.
 */
const Command: FC<CommandProps> = ({ name, defaults, removeMe }) => {
	const { fields, append, remove } = useFieldArray({ name: name });

	// biome-ignore lint/correctness/useExhaustiveDependencies: append stable.
	const addItem = useCallback(() => {
		append({ arg: '' }, { shouldFocus: false });
	}, [name]);
	// Remove the last arg, or the command if only 1 arg.
	// biome-ignore lint/correctness/useExhaustiveDependencies: remove stable.
	const removeLast = useCallback(() => {
		if (fields.length === 1 && removeMe !== undefined) {
			removeMe();
			return;
		}

		remove(fields.length - 1);
	}, [fields.length, removeMe]);

	const placeholder = (index: number) => {
		if (index === 0) return `e.g. "/bin/bash"`;
		if (index === 1) return `e.g. "/opt/script.sh"`;
		return `e.g. "-arg${index - 1}"`;
	};

	return (
		<div className="col-span-full grid grid-cols-subgrid">
			<FieldGroup className="col-span-full grid grid-cols-subgrid gap-2">
				{fields.map(({ id }, argIndex) => (
					<FieldText
						className="py-0"
						key={id}
						name={`${name}.${argIndex}.arg`}
						placeholder={defaults?.[argIndex]?.arg ?? placeholder(argIndex)}
						required
					/>
				))}
			</FieldGroup>

			<div className="col-span-full flex flex-row items-center pt-4">
				{removeMe && (
					<Button
						aria-label="Delete this command"
						onClick={removeMe}
						variant="outline"
					>
						<Trash2 />
					</Button>
				)}
				<ButtonGroup className="ml-auto">
					<Button
						aria-label="Add an argument"
						onClick={addItem}
						size="icon-xs"
						variant="ghost"
					>
						<Plus />
					</Button>
					<Button
						aria-label="Remove the last argument"
						disabled={isEmptyArray(fields)}
						onClick={removeLast}
						size="icon-xs"
						variant="ghost"
					>
						<Minus />
					</Button>
				</ButtonGroup>
			</div>
		</div>
	);
};

export default Command;
