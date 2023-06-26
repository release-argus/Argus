import { FC, useEffect, useMemo, useState } from "react";
import { FormGroup, Stack } from "react-bootstrap";
import {
  ServiceEditAPIType,
  ServiceEditOtherData,
  ServiceEditType,
} from "types/service-edit";

import EditServiceCommands from "components/modals/service-edit/commands";
import EditServiceDashboard from "components/modals/service-edit/dashboard";
import EditServiceDeployedVersion from "components/modals/service-edit/deployed-version";
import EditServiceLatestVersion from "components/modals/service-edit/latest-version";
import EditServiceNotifys from "components/modals/service-edit/notifys";
import EditServiceOptions from "components/modals/service-edit/options";
import EditServiceWebHooks from "components/modals/service-edit/webhooks";
import { FormItem } from "components/generic/form";
import { Loading } from "./loading";
import { WebHookType } from "types/config";
import { convertAPIServiceDataEditToUI } from "./util/api-ui-conversions";
import { fetchJSON } from "utils";
import { useFormContext } from "react-hook-form";
import { useQuery } from "@tanstack/react-query";
import { useWebSocket } from "contexts/websocket";

interface Props {
  name: string;
}

const EditService: FC<Props> = ({ name }) => {
  const { reset } = useFormContext();
  const [loading, setLoading] = useState(true);

  const { data: otherOptionsData, isFetched: isFetchedOtherOptionsData } =
    useQuery(["service/edit", "detail"], () =>
      fetchJSON<ServiceEditOtherData>("api/v1/service/edit")
    );
  const { data: serviceData, isSuccess: isSuccessServiceData } = useQuery(
    ["service/edit", { id: name }],
    () => fetchJSON<ServiceEditAPIType>(`api/v1/service/edit/${name}`),
    {
      enabled: !!name,
      refetchOnMount: "always",
    }
  );

  const defaultData: ServiceEditType = useMemo(
    () => convertAPIServiceDataEditToUI(name, serviceData, otherOptionsData),
    [serviceData, otherOptionsData]
  );
  const { monitorData } = useWebSocket();

  useEffect(() => {
    // If we're loading and have finished fetching the service data
    // (or don't have name = resetting for close)
    if (
      (loading && isSuccessServiceData && isFetchedOtherOptionsData) ||
      !name
    ) {
      reset(defaultData);
      setTimeout(() => setLoading(false), 100);
    }
  }, [defaultData]);

  return loading ? (
    <Loading name={name} />
  ) : (
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
                validation ||
                (value === "" ? "Required" : "name should be unique")
              );
            },
          }}
          col_sm={12}
          label="Name"
          onRight
        />
        <FormItem name="comment" col_sm={12} label="Comment" onRight />
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
        globals={otherOptionsData?.webhook}
        defaults={otherOptionsData?.defaults?.webhook as WebHookType}
        hard_defaults={otherOptionsData?.hard_defaults?.webhook as WebHookType}
      />
      <EditServiceNotifys
        globals={otherOptionsData?.notify}
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
