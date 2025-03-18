import { DeployedVersionLookupType, ServiceOptionsType } from 'types/config';
import { FC, memo, useEffect } from 'react';
import { useFormContext, useWatch } from 'react-hook-form';

import { Accordion } from 'react-bootstrap';
import { DeployedVersionLookupEditType } from 'types/service-edit';
import DeployedVersionManual from './deployed-version-manual';
import DeployedVersionURL from './deployed-version-url';
import { FormSelect } from 'components/generic/form';
import { useWebSocket } from 'contexts/websocket';

interface Props {
	serviceID: string;
	original?: DeployedVersionLookupEditType;
	original_options?: ServiceOptionsType;
	defaults?: DeployedVersionLookupType;
	hard_defaults?: DeployedVersionLookupType;
}

const deployedVersionTypeOptions: {
	label: string;
	value: NonNullable<DeployedVersionLookupType['type']>;
}[] = [
	{ label: 'URL', value: 'url' },
	{ label: 'Manual', value: 'manual' },
];

/**
 * The `deployed_version` form fields.
 *
 * @param serviceID - The name of the service.
 * @param original - The original values of the form.
 * @param original_options - The original service.options of the form.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default.
 * @returns The form fields for the `deployed_version`.
 */
const EditServiceDeployedVersion: FC<Props> = ({
	serviceID,
	original,
	original_options,
	defaults,
	hard_defaults,
}) => {
	const { setValue } = useFormContext();
	const selectedType: string = useWatch({
		name: 'deployed_version.type',
	});
	const { monitorData } = useWebSocket();
	const serviceStatus = monitorData.service?.[serviceID]?.status;
	useEffect(() => {
		if (selectedType === 'manual') {
			setValue(
				'deployed_version.version',
				serviceStatus?.deployed_version ?? serviceStatus?.latest_version ?? '',
			);
		}
	}, [selectedType]);

	return (
		<Accordion>
			<Accordion.Header>Deployed Version:</Accordion.Header>
			<Accordion.Body className="d-flex flex-wrap">
				<FormSelect
					name="deployed_version.type"
					col_xs={6}
					col_lg={2}
					label="Type"
					options={deployedVersionTypeOptions}
				/>
				{selectedType === 'manual' ? (
					<DeployedVersionManual
						serviceID={serviceID}
						original={original}
						original_options={original_options}
					/>
				) : (
					<DeployedVersionURL
						serviceID={serviceID}
						original={original}
						original_options={original_options}
						defaults={defaults}
						hard_defaults={hard_defaults}
					/>
				)}
			</Accordion.Body>
		</Accordion>
	);
};

export default memo(EditServiceDeployedVersion);
