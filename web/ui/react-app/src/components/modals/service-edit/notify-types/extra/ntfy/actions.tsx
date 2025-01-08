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
import { isEmptyArray, isEmptyOrNull } from 'utils';
import { useFieldArray, useFormContext, useWatch } from 'react-hook-form';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { FormLabel } from 'components/generic/form';
import { NotifyNtfyAction } from 'types/config';
import NtfyAction from './action';
import { convertNtfyActionsFromString } from 'components/modals/service-edit/util';
import { diffObjects } from 'utils/diff-objects';

interface Props {
	name: string;
	label: string;
	tooltip: string;
	defaults?: NotifyNtfyAction[];
}

/**
 * NtfyActions returns the form fields for the Ntfy actions.
 *
 * @param name - The name of the field in the form.
 * @param label - The label for the field.
 * @param tooltip - The tooltip for the field.
 * @param defaults - The default values for the field.
 * @returns A set of form fields for a list of Ntfy actions.
 */
const NtfyActions: FC<Props> = ({ name, label, tooltip, defaults }) => {
	const { setValue, trigger } = useFormContext();
	const { fields, append, remove } = useFieldArray({
		name: name,
	});
	const addItem = useCallback(() => {
		append(
			{
				action: 'view',
				label: '',
				url: '',
				intent: 'io.heckel.ntfy.USER_ACTION',
			},
			{ shouldFocus: false },
		);
	}, []);

	// keep track of the array values, so we can switch to defaults when unchanged.
	const fieldValues: NotifyNtfyAction[] = useWatch({ name: name });
	// useDefaults when fieldValues unset, or match the defaults.
	const useDefaults = useMemo(
		() =>
			isEmptyArray(defaults)
				? false
				: !diffObjects(fieldValues, defaults, ['.action']),
		[fieldValues, defaults],
	);

	// Keep only selects/length of arrays.
	const trimmedDefaults = useMemo(
		() => convertNtfyActionsFromString(undefined, JSON.stringify(defaults)),
		[defaults],
	);
	// trigger validation on change of defaults used/not.
	useEffect(() => {
		trigger(name);

		// Give defaults back if field empty.
		if (isEmptyArray(fieldValues)) {
			trimmedDefaults.forEach((value) => {
				append(value, { shouldFocus: false });
			});
		}
	}, [useDefaults]);

	// on load, ensure we don't have another types actions,
	// and give the defaults if not overridden.
	useEffect(() => {
		// ensure we don't have another types actions.
		for (const item of fieldValues ?? []) {
			if (isEmptyOrNull(item.action)) {
				setValue(name, []);
				break;
			}
		}
	}, []);

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
						<NtfyAction
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

export default memo(NtfyActions);
