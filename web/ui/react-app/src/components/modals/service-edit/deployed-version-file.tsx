import { FieldText } from '@/components/generic/field';
import VersionWithRefresh from '@/components/modals/service-edit/version-with-refresh';

/**
 * The `deployed_version` form fields for 'file' version.
 */
const DeployedVersionFile = () => {
	const name = 'deployed_version';

	return (
		<>
			<FieldText
				colSize={{ lg: 6, sm: 12 }}
				label="Path"
				name={`${name}.path`}
				required
				tooltip={{
					content:
						'Absolute path to a file allowed by configured settings.data.file_read_paths',
					type: 'string',
				}}
			/>
			<FieldText
				colSize={{ lg: 6, sm: 12 }}
				label="RegEx"
				name={`${name}.regex`}
				required
				tooltip={{
					content:
						'RegEx to extract the version from the file; include a capture group, e.g. productversion="([^"]+)"',
					type: 'string',
				}}
			/>
			<VersionWithRefresh vType="deployed_version" />
		</>
	);
};

export default DeployedVersionFile;
