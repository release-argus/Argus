import { Minus, Plus } from 'lucide-react';
import { useCallback } from 'react';
import { useFieldArray } from 'react-hook-form';
import { FieldLabel } from '@/components/generic/field';
import FormURLCommand from '@/components/modals/service-edit/latest-version-urlcommand';
import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import { isEmptyArray } from '@/utils';
import { latestVersionURLCommandTypeOptions } from '@/utils/api/types/config/service/latest-version';

/**
 * @returns The form fields for a list of `latest_version.url_commands`.
 */
const FormURLCommands = () => {
	const id = 'latest_version.url_commands';
	const { fields, append, remove } = useFieldArray({ name: id });

	// biome-ignore lint/correctness/useExhaustiveDependencies: append stable.
	const addItem = useCallback(() => {
		append(
			{
				new: '',
				old: '',
				regex: '',
				text: '',
				type: Object.values(latestVersionURLCommandTypeOptions)[0].value,
			},
			{ shouldFocus: false },
		);
	}, []);

	// Remove item at given index.
	// biome-ignore lint/correctness/useExhaustiveDependencies: remove stable.
	const removeItem = useCallback(
		(index: number) => () => {
			document.getElementById(id)?.focus();
			remove(index);
		},
		[],
	);
	// Remove last item.
	// biome-ignore lint/correctness/useExhaustiveDependencies: removeItem stable.
	const removeLast = useCallback(
		() => () => removeItem(fields.length - 1)(),
		[fields.length],
	);

	return (
		<>
			<div className="col-span-full grid grid-cols-subgrid space-y-1">
				<div className="col-span-full flex w-full items-center justify-between">
					<FieldLabel text="URL Commands" />
					<ButtonGroup>
						<Button
							aria-label="Add new URL Command"
							id={id}
							onClick={addItem}
							size="icon-xs"
							variant="ghost"
						>
							<Plus />
						</Button>
						<Button
							aria-label="Remove last URL Command"
							disabled={isEmptyArray(fields)}
							onClick={removeLast()}
							size="icon-xs"
							variant="ghost"
						>
							<Minus />
						</Button>
					</ButtonGroup>
				</div>
			</div>
			<div className="col-span-full grid grid-cols-subgrid gap-2">
				{fields.map(({ id }, i) => (
					<FormURLCommand
						key={id}
						name={`latest_version.url_commands.${i}`}
						removeMe={removeItem(i)}
					/>
				))}
			</div>
			{!isEmptyArray(fields) && <br />}
		</>
	);
};

export default FormURLCommands;
