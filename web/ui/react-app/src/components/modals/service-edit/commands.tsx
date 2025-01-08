import { Accordion, Button, FormGroup, Row } from 'react-bootstrap';
import { FC, memo, useCallback } from 'react';

import Command from './command';
import { isEmptyArray } from 'utils';
import { useFieldArray } from 'react-hook-form';

interface Props {
	name: string;
	loading: boolean;
}

/**
 * The form fields for all commands in a service.
 *
 * @param name - The name of the field in the form.
 * @param loading - Whether the modal is loading.
 * @returns The set of form fields for a list of `command`.
 */
const EditServiceCommands: FC<Props> = ({ name, loading }) => {
	const { fields, append, remove } = useFieldArray({
		name: name,
	});

	const addItem = useCallback(() => {
		append({ args: [{ arg: '' }] }, { shouldFocus: false });
	}, []);

	return (
		<Accordion>
			<Accordion.Header>Command:</Accordion.Header>
			<Accordion.Body>
				<FormGroup className="mb-2">
					<Row>
						{fields.map(({ id }, index) => (
							<Row key={id}>
								<Command
									name={`${name}.${index}.args`}
									removeMe={() => remove(index)}
								/>
							</Row>
						))}
					</Row>
					<Row>
						<Button
							className={isEmptyArray(fields) ? 'mt-2' : ''}
							variant="secondary"
							onClick={addItem}
							disabled={loading}
						>
							Add Command
						</Button>
					</Row>
				</FormGroup>
			</Accordion.Body>
		</Accordion>
	);
};

export default memo(EditServiceCommands);
