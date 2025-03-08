import { Accordion, FormGroup, Row } from 'react-bootstrap';
import {
	DefaultDockerFilterType,
	DefaultLatestVersionRequireType,
	DockerFilterRegistryType,
	DockerFilterType,
} from 'types/config';
import { FC, memo, useEffect, useMemo } from 'react';
import { FormLabel, FormSelect, FormText } from 'components/generic/form';
import { firstNonDefault, isEmptyOrNull } from 'utils';
import { useFormContext, useWatch } from 'react-hook-form';

import Command from './command';

const DockerRegistryOptions = [
	{ label: 'Docker Hub', value: 'hub' },
	{ label: 'GHCR', value: 'ghcr' },
	{ label: 'Quay', value: 'quay' },
];

type Props = {
	defaults?: DefaultLatestVersionRequireType;
	hard_defaults?: DefaultLatestVersionRequireType;
};

/**
 * The `latest_version.require` form fields.
 *
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for the `latest_version.require`.
 */
const EditServiceLatestVersionRequire: FC<Props> = ({
	defaults,
	hard_defaults,
}) => {
	const { setValue } = useFormContext();

	const defaultDockerRegistry: DefaultDockerFilterType['type'] =
		defaults?.docker?.type ?? hard_defaults?.docker?.type;
	const dockerRegistryOptions = useMemo(() => {
		if (defaultDockerRegistry === undefined) return DockerRegistryOptions;

		const defaultDockerRegistryLabel = DockerRegistryOptions.find(
			(option) =>
				option.value.toLowerCase() === defaultDockerRegistry.toLowerCase(),
		);

		if (defaultDockerRegistryLabel)
			return [
				{
					value: '',
					label: `${defaultDockerRegistryLabel.label} (default)`,
				},
				...DockerRegistryOptions,
			];

		// Unknown default registry, return without this default.
		return DockerRegistryOptions;
	}, [defaultDockerRegistry]);
	const dockerRegistry: DockerFilterType['type'] = useWatch({
		name: 'latest_version.require.docker.type',
	});
	const selectedDockerRegistry: DockerFilterRegistryType | undefined =
		dockerRegistry || defaultDockerRegistry;
	const showUsernameField = selectedDockerRegistry === 'hub';

	const convertedDefaults = useMemo(
		() =>
			selectedDockerRegistry
				? {
						token: firstNonDefault(
							defaults?.docker?.[selectedDockerRegistry]?.token,
							hard_defaults?.docker?.[selectedDockerRegistry]?.token,
						),
						username: firstNonDefault(
							defaults?.docker?.[selectedDockerRegistry]?.username,
							hard_defaults?.docker?.[selectedDockerRegistry]?.username,
						),
				  }
				: {
						token: undefined,
						username: undefined,
				  },
		[selectedDockerRegistry],
	);

	useEffect(() => {
		// Default to Docker Hub if no registry selected, and no default registry.
		if (isEmptyOrNull(selectedDockerRegistry))
			setValue('latest_version.require.docker.type', 'hub');
	}, []);

	return (
		<Accordion style={{ marginBottom: '0.5rem' }}>
			<Accordion.Header>Require:</Accordion.Header>
			<Accordion.Body>
				<Row>
					<FormText
						name="latest_version.require.regex_content"
						col_xs={6}
						label={'RegEx Content'}
						tooltip="GitHub=assets must contain a match, URL=webpage must contain a match"
						isRegex
					/>
					<FormText
						name="latest_version.require.regex_version"
						col_xs={6}
						label={'RegEx Version'}
						tooltip="Version found must match, e.g. exclude '*-beta' versions with '^[0-9.]+$'"
						isRegex
						positionXS="right"
					/>

					<FormGroup className="pt-1">
						<FormLabel
							text="Command"
							tooltip="Command to run before a release is considered usable and notified/shown in the UI"
						/>
						<Command name="latest_version.require.command" />
					</FormGroup>

					<hr />
					<FormLabel text="Docker" />
					<FormSelect
						name="latest_version.require.docker.type"
						col_sm={12}
						label="Type"
						options={dockerRegistryOptions}
					/>
					<FormText
						name="latest_version.require.docker.image"
						col_xs={6}
						label="Image"
					/>
					<FormText
						name="latest_version.require.docker.tag"
						col_xs={6}
						label="Tag"
						positionXS="right"
					/>
					{showUsernameField && (
						<FormText
							key="username"
							name="latest_version.require.docker.username"
							col_sm={4}
							label="Username"
							defaultVal={convertedDefaults.username}
						/>
					)}
					<FormText
						name="latest_version.require.docker.token"
						key="token"
						col_sm={showUsernameField ? 8 : 12}
						label="Token"
						defaultVal={convertedDefaults.token}
						positionXS={showUsernameField ? 'right' : 'left'}
					/>
				</Row>
			</Accordion.Body>
		</Accordion>
	);
};

export default memo(EditServiceLatestVersionRequire);
