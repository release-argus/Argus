import { FormItem, FormLabel } from "components/generic/form";

import FormKeyValMap from "components/modals/service-edit/key-val-map";
import { NotifyOpsGenieType } from "types/config";
import { NotifyOptions } from "./generic";
import OpsGenieTargets from "./opsgenie_extra/targets";
import { useGlobalOrDefault } from "./util";

const OPSGENIE = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyOpsGenieType;
  defaults?: NotifyOpsGenieType;
  hard_defaults?: NotifyOpsGenieType;
}) => (
  <>
    <NotifyOptions
      name={name}
      global={global?.options}
      defaults={defaults?.options}
      hard_defaults={hard_defaults?.options}
    />
    <>
      <FormLabel text="URL Fields" heading />
      <FormItem
        name={`${name}.url_fields.host`}
        col_sm={9}
        label="Host"
        tooltip="The OpsGenie API host. Use 'api.eu.opsgenie.com' for EU instances"
        placeholder={useGlobalOrDefault(
          global?.url_fields?.host,
          defaults?.url_fields?.host,
          hard_defaults?.url_fields?.host
        )}
      />
      <FormItem
        name={`${name}.url_fields.port`}
        col_sm={3}
        type="number"
        label="Port"
        placeholder={useGlobalOrDefault(
          global?.url_fields?.port,
          defaults?.url_fields?.port,
          hard_defaults?.url_fields?.port
        )}
        onRight
      />
      <FormItem
        name={`${name}.url_fields.apikey`}
        required
        col_sm={12}
        label="API Key"
        placeholder={useGlobalOrDefault(
          global?.url_fields?.apikey,
          defaults?.url_fields?.apikey,
          hard_defaults?.url_fields?.apikey
        )}
      />
    </>
    <>
      <FormLabel text="Params" heading />
      <FormItem
        name={`${name}.params.actions`}
        label="Actions"
        tooltip="Custom actions that will be available for the alert"
        placeholder={useGlobalOrDefault(
          global?.params?.actions,
          defaults?.params?.actions,
          hard_defaults?.params?.actions
        )}
      />
      <FormItem
        name={`${name}.params.alias`}
        label="Alias"
        tooltip="Client-defined identifier of the alert"
        placeholder={useGlobalOrDefault(
          global?.params?.alias,
          defaults?.params?.alias,
          hard_defaults?.params?.alias
        )}
        onRight
      />
      <FormItem
        name={`${name}.params.description`}
        label="Description"
        tooltip="Description field of the alert"
        placeholder={useGlobalOrDefault(
          global?.params?.description,
          defaults?.params?.description,
          hard_defaults?.params?.description
        )}
      />
      <FormItem
        name={`${name}.params.note`}
        label="Note"
        tooltip="Additional note that will be added while creating the alert"
        placeholder={useGlobalOrDefault(
          global?.params?.note,
          defaults?.params?.note,
          hard_defaults?.params?.note
        )}
        onRight
      />
      <FormKeyValMap
        name={`${name}.params.details`}
        label="Details"
        tooltip="Map of key-val custom props of the alert"
        keyPlaceholder="<key>"
        valuePlaceholder="<value>"
      />
      <FormItem
        name={`${name}.params.entity`}
        label="Entity"
        tooltip="Entity field of the alert that is generally used to specify which domain the Source field of the alert"
        placeholder={useGlobalOrDefault(
          global?.params?.entity,
          defaults?.params?.entity,
          hard_defaults?.params?.entity
        )}
      />
      <FormItem
        name={`${name}.params.priority`}
        type="number"
        label="Priority"
        tooltip="Priority level of the alert. 1/2/3/4/5"
        placeholder={useGlobalOrDefault(
          global?.params?.priority,
          defaults?.params?.priority,
          hard_defaults?.params?.priority
        )}
        onRight
      />
      <OpsGenieTargets
        name={`${name}.params.responders`}
        label="Responders"
        tooltip="Teams, users, escalations and schedules that the alert will be routed to"
      />
      <FormItem
        name={`${name}.params.source`}
        label="Source"
        tooltip="Source field of the alert"
        placeholder={useGlobalOrDefault(
          global?.params?.source,
          defaults?.params?.source,
          hard_defaults?.params?.source
        )}
      />
      <FormItem
        name={`${name}.params.tags`}
        label="Tags"
        tooltip="Tags of the alert"
        placeholder={useGlobalOrDefault(
          global?.params?.tags,
          defaults?.params?.tags,
          hard_defaults?.params?.tags
        )}
        onRight
      />
      <FormItem
        name={`${name}.params.title`}
        label="Title"
        tooltip="Notification title, optionally set by the sender"
        placeholder={useGlobalOrDefault(
          global?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        )}
      />
      <FormItem
        name={`${name}.params.user`}
        label="User"
        tooltip="Display name of the request owner"
        placeholder={useGlobalOrDefault(
          global?.params?.user,
          defaults?.params?.user,
          hard_defaults?.params?.user
        )}
        onRight
      />
    </>
    <OpsGenieTargets
      name={`${name}.params.visibleto`}
      label="Visible To"
      tooltip="Teams and users that the alert will become visible to without sending any notification"
    />
  </>
);

export default OPSGENIE;
