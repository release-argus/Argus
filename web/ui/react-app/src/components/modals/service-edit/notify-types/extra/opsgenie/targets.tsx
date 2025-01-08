import {
	Button,
	ButtonGroup,
	Col,
	FormGroup,
	Row,
	Stack,
} from 'react-bootstrap';
import { FC, memo, useCallback, useEffect, useMemo } from 'react';
import { faMinus, faPlus } from '@fortawesome/free-solid-svg-icons';
import { useFieldArray, useFormContext, useWatch } from 'react-hook-form';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { FormLabel } from 'components/generic/form';
import { NotifyOpsGenieTarget } from 'types/config';
import OpsGenieTarget from './target';
import { diffObjects } from 'utils/diff-objects';
import { isEmptyArray } from 'utils';

interface Props {
	name: string;
	label: string;
	tooltip: string;

	defaults?: NotifyOpsGenieTarget[];
}

/**
 * OpsGenieTargets returns the form fields for a list of OpsGenie targets.
 *
 * @param name - The name of the field in the form.
 * @param label - The label for the field.
 * @param tooltip - The tooltip for the field.
 * @param defaults - The default values for the field.
 * @returns A set of form fields for a list of OpsGenie targets.
 */
const OpsGenieTargets: FC<Props> = ({ name, label, tooltip, defaults }) => {
	const { trigger } = useFormContext();
	const { fields, append, remove } = useFieldArray({
		name: name,
	});
	const addItem = useCallback(() => {
		append(
			{
				type: 'team',
				sub_type: 'id',
				value: '',
			},
			{ shouldFocus: false },
		);
	}, []);

	// keep track of the array values, so can switch to defaults when unchanged.
	const fieldValues: NotifyOpsGenieTarget[] = useWatch({ name: name });
	// useDefaults when fieldValues undefined, or match defaults.
	const useDefaults = useMemo(
		() =>
			isEmptyArray(defaults)
				? false
				: !diffObjects(fieldValues, defaults, ['.type', '.sub_type']),
		[fieldValues, defaults],
	);
	useEffect(() => {
		trigger(name);

		// Give defaults back if field empty.
		if (isEmptyArray(fieldValues))
			defaults?.forEach((value) => {
				append(
					{ type: value.type, sub_type: value.sub_type, value: '' },
					{ shouldFocus: false },
				);
			});
	}, [useDefaults]);

	// Remove the last item if not the only one, or doesn't match the defaults.
	const removeLast = useCallback(() => {
		!(useDefaults && fields.length == 1) && remove(fields.length - 1);
	}, [fields.length, useDefaults]);

	return (
		<FormGroup>
			<Row>
				<Col className="pt-1">
					<FormLabel text={label} tooltip={tooltip} />
				</Col>
				<Col>
					<ButtonGroup style={{ float: 'right' }}>
						<Button
							aria-label={`Add new ${label}`}
							className="btn-unchecked"
							style={{ float: 'right' }}
							onClick={addItem}
						>
							<FontAwesomeIcon icon={faPlus} />
						</Button>
						<Button
							aria-label={`Remove last ${label}`}
							className="btn-unchecked"
							style={{ float: 'left' }}
							onClick={removeLast}
							disabled={isEmptyArray(fields)}
						>
							<FontAwesomeIcon icon={faMinus} />
						</Button>
					</ButtonGroup>
				</Col>
			</Row>
			<Stack>
				{fields.map(({ id }, index) => (
					<Row key={id}>
						<OpsGenieTarget
							name={`${name}.${index}`}
							removeMe={
								// Give the remove that is disabled if only one item, and it matches the defaults.
								fieldValues?.length === 1 ? removeLast : () => remove(index)
							}
							defaults={useDefaults ? defaults?.[index] : undefined}
						/>
					</Row>
				))}
			</Stack>
		</FormGroup>
	);
};

export default memo(OpsGenieTargets);
