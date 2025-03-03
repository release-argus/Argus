import { Button, Col, Row } from 'react-bootstrap';
import { FC, memo } from 'react';
import { FormSelect, FormText } from 'components/generic/form';
import { NotifyNtfyAction, NotifyNtfyActionTypes } from 'types/config';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import RenderAction from './render';
import { faTrash } from '@fortawesome/free-solid-svg-icons';
import { useWatch } from 'react-hook-form';

interface Props {
	name: string;
	defaults?: NotifyNtfyAction;
	removeMe: () => void;
}

/**
 * The form fields for a NTFY action.
 *
 * @param name - The name of the field in the form.
 * @param defaults - The default values for the action.
 * @param removeMe - The function to remove this action.
 * @returns The form fields for this action.
 */
const NtfyAction: FC<Props> = ({ name, defaults, removeMe }) => {
	const typeOptions: { label: string; value: NotifyNtfyActionTypes }[] = [
		{ label: 'View', value: 'view' },
		{ label: 'HTTP', value: 'http' },
		{ label: 'Broadcast', value: 'broadcast' },
	];
	enum typeLabelMap {
		view = 'Open page',
		http = 'Close door',
		broadcast = 'Take picture',
	}

	const targetType: keyof typeof typeLabelMap = useWatch({
		name: `${name}.action`,
	});

	return (
		<>
			<Col xs={2} sm={1} className="py-1 pe-2">
				<Button
					aria-label="Remove this action"
					className="btn-secondary-outlined btn-icon-center p-0"
					variant="secondary"
					onClick={removeMe}
				>
					<FontAwesomeIcon icon={faTrash} />
				</Button>
			</Col>
			<Col xs={10} sm={11}>
				<Row>
					<FormSelect
						name={`${name}.action`}
						col_xs={6}
						label="Action Type"
						options={typeOptions}
					/>
					<FormText
						name={`${name}.label`}
						required
						col_xs={6}
						label="Label"
						tooltip="Button name to display on the notification"
						defaultVal={defaults?.label}
						placeholder={`e.g. '${typeLabelMap[targetType]}'`}
						positionXS="right"
						positionSM="middle"
					/>
					<RenderAction
						name={name}
						targetType={targetType}
						defaults={defaults}
					/>
				</Row>
			</Col>
		</>
	);
};

export default memo(NtfyAction);
