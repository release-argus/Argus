import { ServiceEditOtherData, ServiceEditType } from 'types/service-edit';

import EditServiceCommands from 'components/modals/service-edit/commands';
import EditServiceDashboard from 'components/modals/service-edit/dashboard';
import EditServiceDeployedVersion from 'components/modals/service-edit/deployed-version';
import EditServiceLatestVersion from 'components/modals/service-edit/latest-version';
import EditServiceNotifiers from 'components/modals/service-edit/notifiers';
import EditServiceOptions from 'components/modals/service-edit/options';
import EditServiceRoot from 'components/modals/service-edit/root';
import EditServiceWebHooks from 'components/modals/service-edit/webhooks';
import { FC } from 'react';
import { Stack } from 'react-bootstrap';
import { WebHookType } from 'types/config';
import { useWebSocket } from 'contexts/websocket';

interface Props {
	id: string;
	name?: string;
	defaultData: ServiceEditType;
	otherOptionsData?: ServiceEditOtherData;
	loading: boolean;
}

/**
 * The form fields for creating/editing a service.
 *
 * @param id - The ID of the Service.
 * @param name - The name of the service.
 * @param defaultData - The default data for the service.
 * @param otherOptionsData - The other options data, containing globals/defaults/hardDefaults.
 * @param loading - Whether the modal is loading.
 * @returns The form fields for creating/editing a service.
 */
const EditService: FC<Props> = ({
	id,
	name,
	defaultData,
	otherOptionsData,
	loading,
}) => {
	const { monitorData } = useWebSocket();

	return (
		<Stack gap={3}>
			<EditServiceRoot
				id={id}
				name={name}
				original_name={defaultData?.name}
				loading={loading}
			/>
			<EditServiceOptions
				defaults={otherOptionsData?.defaults?.service?.options}
				hard_defaults={otherOptionsData?.hard_defaults?.service?.options}
			/>
			<EditServiceLatestVersion
				serviceID={id}
				original={defaultData?.latest_version}
				original_options={defaultData?.options}
				defaults={otherOptionsData?.defaults?.service?.latest_version}
				hard_defaults={otherOptionsData?.hard_defaults?.service?.latest_version}
			/>
			<EditServiceDeployedVersion
				serviceID={id}
				original={defaultData?.deployed_version}
				original_options={defaultData?.options}
				defaults={otherOptionsData?.defaults?.service?.deployed_version}
				hard_defaults={
					otherOptionsData?.hard_defaults?.service?.deployed_version
				}
			/>
			<EditServiceCommands name="command" loading={loading} />
			<EditServiceWebHooks
				mains={otherOptionsData?.webhook}
				defaults={otherOptionsData?.defaults?.webhook as WebHookType}
				hard_defaults={otherOptionsData?.hard_defaults?.webhook as WebHookType}
				loading={loading}
			/>
			<EditServiceNotifiers
				serviceID={id}
				originals={defaultData?.notify}
				mains={otherOptionsData?.notify}
				defaults={otherOptionsData?.defaults?.notify}
				hard_defaults={otherOptionsData?.hard_defaults?.notify}
				loading={loading}
			/>
			<EditServiceDashboard
				serviceID={id}
				originals={defaultData?.dashboard}
				defaults={otherOptionsData?.defaults?.service?.dashboard}
				hard_defaults={otherOptionsData?.hard_defaults?.service?.dashboard}
				serviceStatus={monitorData.service?.[id]?.status}
			/>
		</Stack>
	);
};

export default EditService;
