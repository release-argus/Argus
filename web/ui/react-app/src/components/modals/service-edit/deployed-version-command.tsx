import { Minus, Plus } from 'lucide-react';
import { useCallback } from 'react';
import { useFieldArray } from 'react-hook-form';
import { FieldText } from '@/components/generic/field';
import VersionWithRefresh from '@/components/modals/service-edit/version-with-refresh';
import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import { isEmptyArray } from '@/utils';

/**
 * The `deployed_version` form fields for 'command' version.
 */
const DeployedVersionCommand = () => {
	const name = 'deployed_version';
	const commandName = `${name}.command`;
	const { fields, append, remove } = useFieldArray({ name: commandName });

	// biome-ignore lint/correctness/useExhaustiveDependencies: append stable.
	const addArgument = useCallback(() => {
		append('');
	}, [commandName]);
	// Remove the last argument.
	const removeLastArgument = useCallback(() => {
		remove(fields.length - 1);
	}, [fields.length, remove]);

	const placeholder = (index: number) => {
		if (index === 0) return 'e.g. "ssh"';
		return `e.g. "arg${index}"`;
	};

	return (
		<>
			<div className="col-span-full grid grid-cols-subgrid">
				<div className="col-span-full grid grid-cols-subgrid gap-2">
					{fields.map(({ id }, argIndex) => (
						<FieldText
							className="py-0"
							key={id}
							name={`${commandName}.${argIndex}`}
							placeholder={placeholder(argIndex)}
							required
						/>
					))}
				</div>

				<div className="col-span-full flex flex-row items-center pt-4">
					<ButtonGroup className="ml-auto">
						<Button
							aria-label="Add an argument"
							onClick={addArgument}
							size="icon-xs"
							variant="ghost"
						>
							<Plus />
						</Button>
						<Button
							aria-label="Remove the last argument"
							disabled={isEmptyArray(fields)}
							onClick={removeLastArgument}
							size="icon-xs"
							variant="ghost"
						>
							<Minus />
						</Button>
					</ButtonGroup>
				</div>
			</div>
			<FieldText
				colSize={{ sm: 5 }}
				label="RegEx"
				name={`${name}.regex`}
				tooltip={{
					ariaLabel:
						'RegEx to extract the version from the command output, e.g. ([0-9.]+)',
					content: (
						<>
							RegEx to extract the version from the command output, e.g.{' '}
							<span className="bold underline">([0-9.]+)</span>
						</>
					),
					type: 'element',
				}}
			/>
			<VersionWithRefresh className="order-3" vType="deployed_version" />
		</>
	);
};

export default DeployedVersionCommand;