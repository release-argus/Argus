import { Button, Col, Row } from 'react-bootstrap';
import { FC, memo } from 'react';
import { FormSelect, FormText } from 'components/generic/form';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { NotifyOpsGenieTarget } from 'types/config';
import { faTrash } from '@fortawesome/free-solid-svg-icons';
import { useWatch } from 'react-hook-form';

interface Props {
	name: string;
	removeMe: () => void;

	defaults?: NotifyOpsGenieTarget;
}

/**
 * OpsGenieTarget renders fields for an OpsGenie target.
 *
 * @param name - The name of the field in the form.
 * @param removeMe - The function to remove this target.
 * @param defaults - The default values for the target.
 * @returns The form fields for this OpsGenie target.
 */
const OpsGenieTarget: FC<Props> = ({ name, removeMe, defaults }) => {
	const targetTypes: { label: string; value: NotifyOpsGenieTarget['type'] }[] =
		[
			{ label: 'Team', value: 'team' },
			{ label: 'User', value: 'user' },
		];

	const targetType: NotifyOpsGenieTarget['type'] = useWatch({
		name: `${name}.type`,
	});

	return (
		<>
			<Col xs={2} sm={1} style={{ padding: '0.25rem' }}>
				<Button
					className="btn-secondary-outlined btn-icon-center"
					variant="secondary"
					onClick={removeMe}
				>
					<FontAwesomeIcon icon={faTrash} />
				</Button>
			</Col>
			<Col xs={10} sm={11}>
				<Row>
					<FormSelect
						name={`${name}.type`}
						col_xs={6}
						col_sm={3}
						options={targetTypes}
					/>
					<FormSelect
						name={`${name}.sub_type`}
						col_xs={6}
						col_sm={3}
						options={[
							{ label: 'ID', value: 'id' },
							targetType === 'team'
								? { label: 'Name', value: 'name' }
								: { label: 'Username', value: 'username' },
						]}
						position="middle"
						positionXS="right"
					/>
					<FormText
						name={`${name}.value`}
						required
						col_sm={6}
						defaultVal={defaults?.value}
						position="right"
					/>
				</Row>
			</Col>
		</>
	);
};

export default memo(OpsGenieTarget);
