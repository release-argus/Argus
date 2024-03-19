import {
  FormItem,
  FormKeyValMap,
  FormLabel,
  FormList,
} from "components/generic/form";
import {
  convertHeadersFromString,
  convertOpsGenieTargetFromString,
  convertStringToFieldArray,
} from "components/modals/service-edit/util";

import { NotifyOpsGenieType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { OpsGenieTargets } from "components/modals/service-edit/notify-types/extra";
import { globalOrDefault } from "components/modals/service-edit/notify-types/util";
import { useMemo } from "react";

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
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      apiKey: globalOrDefault(
        global?.url_fields?.apikey,
        defaults?.url_fields?.apikey,
        hard_defaults?.url_fields?.apikey
      ),
      host: globalOrDefault(
        global?.url_fields?.host,
        defaults?.url_fields?.host,
        hard_defaults?.url_fields?.host
      ),
      port: globalOrDefault(
        global?.url_fields?.port,
        defaults?.url_fields?.port,
        hard_defaults?.url_fields?.port
      ),
      // Params
      actions: convertStringToFieldArray(
        globalOrDefault(
          global?.params?.actions as string,
          defaults?.params?.actions as string,
          hard_defaults?.params?.actions as string
        )
      ),
      alias: globalOrDefault(
        global?.params?.alias,
        defaults?.params?.alias,
        hard_defaults?.params?.alias
      ),
      description: globalOrDefault(
        global?.params?.description,
        defaults?.params?.description,
        hard_defaults?.params?.description
      ),
      details: convertHeadersFromString(
        globalOrDefault(
          global?.params?.details as string,
          defaults?.params?.details as string,
          hard_defaults?.params?.details as string
        )
      ),
      entity: globalOrDefault(
        global?.params?.entity,
        defaults?.params?.entity,
        hard_defaults?.params?.entity
      ),
      note: globalOrDefault(
        global?.params?.note,
        defaults?.params?.note,
        hard_defaults?.params?.note
      ),
      priority: globalOrDefault(
        global?.params?.priority,
        defaults?.params?.priority,
        hard_defaults?.params?.priority
      ),
      responders: convertOpsGenieTargetFromString(
        globalOrDefault(
          global?.params?.responders as string,
          defaults?.params?.responders as string,
          hard_defaults?.params?.responders as string
        )
      ),
      source: globalOrDefault(
        global?.params?.source,
        defaults?.params?.source,
        hard_defaults?.params?.source
      ),
      tags: globalOrDefault(
        global?.params?.tags,
        defaults?.params?.tags,
        hard_defaults?.params?.tags
      ),
      title: globalOrDefault(
        global?.params?.title,
        defaults?.params?.title,
        hard_defaults?.params?.title
      ),
      user: globalOrDefault(
        global?.params?.user,
        defaults?.params?.user,
        hard_defaults?.params?.user
      ),
      visibleto: convertOpsGenieTargetFromString(
        globalOrDefault(
          global?.params?.visibleto as string,
          defaults?.params?.visibleto as string,
          hard_defaults?.params?.visibleto as string
        )
      ),
    }),
    [global, defaults, hard_defaults]
  );

  return (
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
          defaultVal={convertedDefaults.host}
        />
        <FormItem
          name={`${name}.url_fields.port`}
          col_sm={3}
          type="number"
          label="Port"
          defaultVal={convertedDefaults.port}
          onRight
        />
        <FormItem
          name={`${name}.url_fields.apikey`}
          required
          col_sm={12}
          label="API Key"
          defaultVal={convertedDefaults.apiKey}
        />
      </>
      <>
        <FormLabel text="Params" heading />
        <FormList
          name={`${name}.params.actions`}
          label="Actions"
          tooltip="Custom actions that will be available for the alert"
          defaults={convertedDefaults.actions}
        />
        <FormItem
          name={`${name}.params.alias`}
          label="Alias"
          tooltip="Client-defined identifier of the alert"
          defaultVal={convertedDefaults.alias}
        />
        <FormItem
          name={`${name}.params.description`}
          label="Description"
          tooltip="Description field of the alert"
          defaultVal={convertedDefaults.description}
          onRight
        />
        <FormItem
          name={`${name}.params.note`}
          label="Note"
          tooltip="Additional note that will be added while creating the alert"
          defaultVal={convertedDefaults.note}
        />
        <FormKeyValMap
          name={`${name}.params.details`}
          label="Details"
          tooltip="Map of key-val custom props of the alert"
          keyPlaceholder="e.g. X-Authorization"
          valuePlaceholder="e.g. 'Bearer TOKEN'"
          defaults={convertedDefaults.details}
        />
        <FormItem
          name={`${name}.params.entity`}
          label="Entity"
          tooltip="Entity field of the alert that is generally used to specify which domain the Source field of the alert"
          defaultVal={convertedDefaults.entity}
        />
        <FormItem
          name={`${name}.params.priority`}
          type="number"
          label="Priority"
          tooltip="Priority level of the alert. 1/2/3/4/5"
          defaultVal={convertedDefaults.priority}
          onRight
        />
        <OpsGenieTargets
          name={`${name}.params.responders`}
          label="Responders"
          tooltip="Teams, users, escalations and schedules that the alert will be routed to"
          defaults={convertedDefaults.responders}
        />
        <FormItem
          name={`${name}.params.source`}
          label="Source"
          tooltip="Source field of the alert"
          defaultVal={convertedDefaults.source}
        />
        <FormItem
          name={`${name}.params.tags`}
          label="Tags"
          tooltip="Tags of the alert"
          defaultVal={convertedDefaults.tags}
          onRight
        />
        <FormItem
          name={`${name}.params.title`}
          label="Title"
          tooltip="Notification title, optionally set by the sender"
          defaultVal={convertedDefaults.title}
        />
        <FormItem
          name={`${name}.params.user`}
          label="User"
          tooltip="Display name of the request owner"
          defaultVal={convertedDefaults.user}
          onRight
        />
      </>
      <OpsGenieTargets
        name={`${name}.params.visibleto`}
        label="Visible To"
        tooltip="Teams and users that the alert will become visible to without sending any notification"
        defaults={convertedDefaults.visibleto}
      />
    </>
  );
};

export default OPSGENIE;
