import type { FC } from 'react';
import EditServiceCommands from '@/components/modals/service-edit/commands';
import EditServiceDashboard from '@/components/modals/service-edit/dashboard';
import EditServiceDeployedVersion from '@/components/modals/service-edit/deployed-version';
import EditServiceLatestVersion from '@/components/modals/service-edit/latest-version';
import EditServiceNotifiers from '@/components/modals/service-edit/notifiers';
import EditServiceOptions from '@/components/modals/service-edit/options';
import EditServiceRoot from '@/components/modals/service-edit/root';
import EditServiceWebHooks from '@/components/modals/service-edit/webhooks';
import { Accordion } from '@/components/ui/accordion';

type EditServiceProps = {
	/* Indicates whether the modal shows a loading state. */
	loading: boolean;
};

/**
 * The form fields for creating/editing a service.
 *
 * @param id - The ID of the Service.
 * @param loading - Indicates whether the modal shows a loading state.
 * @returns The form fields for creating/editing a service.
 */
const EditService: FC<EditServiceProps> = ({ loading }) => {
	return (
		<div className="flex flex-col gap-4">
			<EditServiceRoot loading={loading} />
			<Accordion className="flex flex-col gap-2" type="multiple">
				<EditServiceOptions />
				<EditServiceLatestVersion />
				<EditServiceDeployedVersion />
				<EditServiceCommands loading={loading} name="command" />
				<EditServiceWebHooks loading={loading} />
				<EditServiceNotifiers loading={loading} />
				<EditServiceDashboard />
			</Accordion>
		</div>
	);
};

export default EditService;
