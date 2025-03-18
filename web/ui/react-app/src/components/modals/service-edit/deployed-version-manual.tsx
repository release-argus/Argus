import {
	Alert,
	Button,
	Col,
	FormControl,
	FormGroup,
	InputGroup,
} from 'react-bootstrap';
import { FC, memo, useState } from 'react';
import { NonNullable, ServiceOptionsType } from 'types/config';
import { beautifyGoErrors, fetchVersionJSON } from 'utils';
import { faSave, faSpinner } from '@fortawesome/free-solid-svg-icons';
import { useFormContext, useWatch } from 'react-hook-form';

import { DeployedVersionLookupEditType } from 'types/service-edit';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { FormLabel } from 'components/generic/form';
import { StatusSummaryType } from 'types/summary';
import { TooltipWithAriaProps } from 'components/generic/tooltip';
import cx from 'classnames';
import { formPadding } from 'components/generic/form-shared';
import { useQuery } from '@tanstack/react-query';
import useValuesRefetch from 'hooks/values-refetch';
import { useWebSocket } from 'contexts/websocket';

interface Props {
	serviceID: string;
	original?: DeployedVersionLookupEditType;
	original_options?: ServiceOptionsType;
}

type DeployedVersionManualProps = TooltipWithAriaProps & Props;

/**
 * The `deployed_version` form fields.
 *
 * @param serviceID - The name of the service.
 * @param original - The original values of the form.
 * @param original_options - The original service.options of the form.
 * @returns The form fields for the `deployed_version`.
 */
const DeployedVersionManual: FC<DeployedVersionManualProps> = ({
	serviceID,
	original,
	original_options,
}) => {
	const name = 'deployed_version.version';
	const { register } = useFormContext();
	const [lastFetched, setLastFetched] = useState(0);
	const value: string = useWatch({ name: name });
	const { data: semanticVersioning, refetchData: refetchSemanticVersioning } =
		useValuesRefetch<boolean>('options.semantic_versioning');
	const { monitorData } = useWebSocket();
	const status = (monitorData.service[serviceID]?.status ??
		{}) as NonNullable<StatusSummaryType>;

	const canSave =
		original?.type === 'manual' && value !== status.deployed_version;

	const sizing = {
		col_xs: 12,
		col_sm: 6,
		col_lg: 10,
	};
	const padding = formPadding({
		...sizing,
		positionSM: 'right',
	});

	const handleSave = async () => {
		// Prevent refetching too often.
		const currentTime = Date.now();
		if (currentTime - lastFetched < 1000) return;

		refetchSemanticVersioning();
		// setTimeout to allow time for refetch setStates ^
		const timeout = setTimeout(() => {
			if (value) {
				saveVersion().then((data) => {
					if (data.data) status.deployed_version = data.data?.version;
				});
				setLastFetched(currentTime);
			}
		});
		return () => clearTimeout(timeout);
	};

	const {
		data: versionData,
		isFetching: isSaving,
		refetch: saveVersion,
	} = useQuery({
		queryKey: [
			'version/refresh',
			'deployed_version',
			{ id: serviceID },
			{
				params: { version: value },
				semantic_versioning: semanticVersioning,
			},
		],
		queryFn: () =>
			fetchVersionJSON({
				serviceID,
				dataTarget: 'deployed_version',
				semanticVersioning,
				options: original_options,
				data: { type: 'manual', version: value },
				original: { type: original?.type },
			}),
		enabled: false,
		initialData: {
			version: status.deployed_version ?? '',
			error: '',
			timestamp: '',
		},
		notifyOnChangeProps: 'all',
		staleTime: 0,
	});
	const errorMessage = versionData.error || versionData.message;

	return (
		<>
			<Col
				xs={sizing.col_xs}
				sm={sizing.col_sm}
				lg={sizing.col_lg}
				className={`${padding} pt-1 pb-1 col-form`}
			>
				<FormGroup>
					<FormLabel
						htmlFor={name}
						text="Version"
						tooltip="The version that you have deployed"
					/>
					<InputGroup className="me-3">
						<FormControl
							id={name}
							aria-label="Version string"
							aria-describedby={cx(errorMessage && `${name}-error`)}
							type="text"
							defaultValue={value}
							isInvalid={!!errorMessage}
							{...register(name)}
						/>
						{canSave && value && (
							<Button
								aria-label="Save version"
								variant="secondary"
								className="curved-right-only"
								onClick={handleSave}
								disabled={isSaving || !value || !!errorMessage}
							>
								<FontAwesomeIcon
									icon={isSaving ? faSpinner : faSave}
									spin={isSaving}
								/>
							</Button>
						)}
					</InputGroup>
				</FormGroup>
			</Col>
			{errorMessage && (
				<span
					className="mb-2 pt-1"
					style={{ width: '100%', wordBreak: 'break-all' }}
				>
					<Alert variant="danger">
						Failed to change version:
						<br />
						{beautifyGoErrors(errorMessage)}
					</Alert>
				</span>
			)}
		</>
	);
};

export default memo(DeployedVersionManual);
