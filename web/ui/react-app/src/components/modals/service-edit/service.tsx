import { FormGroup, Stack } from "react-bootstrap";
import { ServiceEditOtherData, ServiceEditType } from "types/service-edit";

import EditServiceCommands from "components/modals/service-edit/commands";
import EditServiceDashboard from "components/modals/service-edit/dashboard";
import EditServiceDeployedVersion from "components/modals/service-edit/deployed-version";
import EditServiceLatestVersion from "components/modals/service-edit/latest-version";
import EditServiceNotifys from "components/modals/service-edit/notifys";
import EditServiceOptions from "components/modals/service-edit/options";
import EditServiceWebHooks from "components/modals/service-edit/webhooks";
import { FC } from "react";
import { FormItem } from "components/generic/form";
import { WebHookType } from "types/config";
import { useWebSocket } from "contexts/websocket";

interface Props {
  name: string;
  defaultData: ServiceEditType;
  otherOptionsData: ServiceEditOtherData;
}

/**
 * Returns the form fields for creating/editing a service
 *
 * @param name - The name of the service
 * @returns The form fields for creating/editing a service
 */
const EditService: FC<Props> = ({ name, defaultData, otherOptionsData }) => {
  const { monitorData } = useWebSocket();

  return (
    <Stack gap={3}>
      <FormGroup className="mb-2">
        <FormItem
          name="name"
          required
          registerParams={{
            validate: (value: string) => {
              const validation =
                value === ""
                  ? false
                  : // Name hasn't changed or name isn't in use
                    name === value || !monitorData.order.includes(value);
              return (
                validation || (value === "" ? "Required" : "Must be unique")
              );
            },
          }}
          col_sm={12}
          label="Name"
        />
        <FormItem name="comment" col_sm={12} label="Comment" position="right" />
      </FormGroup>
      <EditServiceOptions
        defaults={otherOptionsData?.defaults?.service?.options}
        hard_defaults={otherOptionsData?.hard_defaults?.service?.options}
      />
      <EditServiceLatestVersion
        serviceName={name}
        original={defaultData?.latest_version}
        defaults={otherOptionsData?.defaults?.service?.latest_version}
        hard_defaults={otherOptionsData?.hard_defaults?.service?.latest_version}
      />
      <EditServiceDeployedVersion
        serviceName={name}
        original={defaultData?.deployed_version}
        defaults={otherOptionsData?.defaults?.service?.deployed_version}
        hard_defaults={
          otherOptionsData?.hard_defaults?.service?.deployed_version
        }
      />
      <EditServiceCommands name="command" />
      <EditServiceWebHooks
        mains={otherOptionsData?.webhook}
        defaults={otherOptionsData?.defaults?.webhook as WebHookType}
        hard_defaults={otherOptionsData?.hard_defaults?.webhook as WebHookType}
      />
      <EditServiceNotifys
        serviceName={name}
        originals={defaultData?.notify}
        mains={otherOptionsData?.notify}
        defaults={otherOptionsData?.defaults?.notify}
        hard_defaults={otherOptionsData?.hard_defaults?.notify}
      />
      <EditServiceDashboard
        defaults={otherOptionsData?.defaults?.service?.dashboard}
        hard_defaults={otherOptionsData?.hard_defaults?.service?.dashboard}
      />
    </Stack>
  );
};

export default EditService;
