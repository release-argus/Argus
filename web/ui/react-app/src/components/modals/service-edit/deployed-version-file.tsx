import { useWatch } from 'react-hook-form';
import { FieldText } from '@/components/generic/field';
import VersionWithRefresh from '@/components/modals/service-edit/version-with-refresh';

/**
 * The `deployed_version` form fields for 'file' version.
 */
const DeployedVersionFile = () => {
	const name = 'deployed_version';
	const regexTemplate = (useWatch({
		name: `${name}.regex_template`,
	}) ?? '') as string;

	return (
		<>
			<FieldText
				colSize={{ lg: 9, sm: 8, xs: 8 }}
				label="Path"
				name={`${name}.path`}
				tooltip={{
					content:
						'Absolute path to a file allowed by configured settings.data.file_read_paths',
					type: 'string',
				}}
			/>
			<FieldText
				colSize={{ xs: 6 }}
				label="RegEx"
				name={`${name}.regex`}
				required={regexTemplate ? 'Required for templating' : undefined}
				tooltip={{
					content:
						'Optional RegEx to extract the version from the file. Without one, the full file contents are used.',
					type: 'string',
				}}
			/>
			<FieldText
				colSize={{ xs: 6 }}
				label="RegEx Template"
				name={`${name}.regex_template`}
				tooltip={{
					content:
						'Optional template for the RegEx match; e.g. $1.$2.$3 uses the first three capture groups.',
					type: 'string',
				}}
			/>
			<VersionWithRefresh vType="deployed_version" />
		</>
	);
};

export default DeployedVersionFile;
