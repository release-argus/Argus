import { useFormContext, useWatch } from 'react-hook-form';
import { FieldSelect } from '@/components/generic/field';
import DeployedVersionCommand from '@/components/modals/service-edit/deployed-version-command';
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
	const { serviceID } = useSchemaContext();
	const { setValue } = useFormContext();

	const selectedType = useWatch({
		name: `${name}.type`,
	}) as DeployedVersionLookupType;
	const { data: serviceData } = useServiceSummary(serviceID);

	return (
		<AccordionItem value="deployed_version">
			<AccordionTrigger>Deployed Version:</AccordionTrigger>
			<AccordionContent className="grid grid-cols-12 gap-2">
				<FieldSelect
					colSize={{ lg: 2, xs: 6 }}
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
				{selectedType === DEPLOYED_VERSION_LOOKUP_TYPE.COMMAND.value ? (
					<DeployedVersionCommand />
				) : selectedType === DEPLOYED_VERSION_LOOKUP_TYPE.URL.value ? (
					<DeployedVersionURL />
				) : (
					<DeployedVersionManual />
				)}
			</AccordionContent>
		</AccordionItem>
	);
};

export default EditServiceDeployedVersion;
