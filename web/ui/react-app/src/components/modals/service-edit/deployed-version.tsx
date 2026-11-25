import { useFormContext, useWatch } from 'react-hook-form';
import { FieldSelect } from '@/components/generic/field';
import DeployedVersionFile from '@/components/modals/service-edit/deployed-version-file';
import DeployedVersionManual from '@/components/modals/service-edit/deployed-version-manual';
import DeployedVersionURL from '@/components/modals/service-edit/deployed-version-url';
import {
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from '@/components/ui/accordion';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import { useServiceSummary } from '@/hooks/use-service-summary';
import {
	DEPLOYED_VERSION_LOOKUP_TYPE,
	type DeployedVersionLookupType,
	deployedVersionLookupTypeOptions,
} from '@/utils/api/types/config/service/deployed-version';

/**
 * The `deployed_version` form fields.
 */
const EditServiceDeployedVersion = () => {
	const name = 'deployed_version';
	const { serviceID, schemaDataDefaults } = useSchemaContext();
	const defaultType = schemaDataDefaults?.deployed_version?.type;
	const { setValue } = useFormContext();

	const selectedType =
		(useWatch({
			name: `${name}.type`,
		}) as DeployedVersionLookupType | '') ||
		(defaultType as DeployedVersionLookupType);
	const { data: serviceData } = useServiceSummary(serviceID);

	// URL type: fill the row with type+method.
	const typeColSize =
		selectedType === DEPLOYED_VERSION_LOOKUP_TYPE.URL.value ? 6 : 4;

	return (
		<AccordionItem value="deployed_version">
			<AccordionTrigger>Deployed Version:</AccordionTrigger>
			<AccordionContent className="grid grid-cols-12 gap-2">
				<FieldSelect
					colSize={{ lg: 3, sm: typeColSize, xs: typeColSize }}
					label="Type"
					name={`${name}.type`}
					onChange={(opt) => {
						const serviceStatus = serviceData?.status;
						if (opt?.value === DEPLOYED_VERSION_LOOKUP_TYPE.MANUAL.value) {
							setValue(
								`${name}.version`,
								serviceStatus?.deployed_version ??
									serviceStatus?.latest_version ??
									'',
							);
						}
						return opt;
					}}
					options={deployedVersionLookupTypeOptions}
				/>
				{selectedType === DEPLOYED_VERSION_LOOKUP_TYPE.FILE.value ? (
					<DeployedVersionFile />
				) : selectedType === DEPLOYED_VERSION_LOOKUP_TYPE.MANUAL.value ? (
					<DeployedVersionManual />
				) : (
					<DeployedVersionURL />
				)}
			</AccordionContent>
		</AccordionItem>
	);
};

export default EditServiceDeployedVersion;
