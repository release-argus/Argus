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
import { firstNonDefault } from "utils";
import { useMemo } from "react";

/**
 * Returns the form fields for `OpsGenie`
 *
 * @param name - The path to this `OpsGenie` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `OpsGenie` `Notify`
 */
const OPSGENIE = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyOpsGenieType;
  defaults?: NotifyOpsGenieType;
  hard_defaults?: NotifyOpsGenieType;
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        apiKey: firstNonDefault(
          main?.url_fields?.apikey,
          defaults?.url_fields?.apikey,
          hard_defaults?.url_fields?.apikey
        ),
        host: firstNonDefault(
          main?.url_fields?.host,
          defaults?.url_fields?.host,
          hard_defaults?.url_fields?.host
        ),
        port: firstNonDefault(
          main?.url_fields?.port,
          defaults?.url_fields?.port,
          hard_defaults?.url_fields?.port
        ),
      },
      // Params
      params: {
        actions: convertStringToFieldArray(
          firstNonDefault(
            main?.params?.actions as string,
            defaults?.params?.actions as string,
            hard_defaults?.params?.actions as string
          )
        ),
        alias: firstNonDefault(
          main?.params?.alias,
          defaults?.params?.alias,
          hard_defaults?.params?.alias
        ),
        description: firstNonDefault(
          main?.params?.description,
          defaults?.params?.description,
          hard_defaults?.params?.description
        ),
        details: convertHeadersFromString(
          firstNonDefault(
            main?.params?.details as string,
            defaults?.params?.details as string,
            hard_defaults?.params?.details as string
          )
        ),
        entity: firstNonDefault(
          main?.params?.entity,
          defaults?.params?.entity,
          hard_defaults?.params?.entity
        ),
        note: firstNonDefault(
          main?.params?.note,
          defaults?.params?.note,
          hard_defaults?.params?.note
        ),
        priority: firstNonDefault(
          main?.params?.priority,
          defaults?.params?.priority,
          hard_defaults?.params?.priority
        ),
        responders: convertOpsGenieTargetFromString(
          firstNonDefault(
            main?.params?.responders as string,
            defaults?.params?.responders as string,
            hard_defaults?.params?.responders as string
          )
        ),
        source: firstNonDefault(
          main?.params?.source,
          defaults?.params?.source,
          hard_defaults?.params?.source
        ),
        tags: firstNonDefault(
          main?.params?.tags,
          defaults?.params?.tags,
          hard_defaults?.params?.tags
        ),
        title: firstNonDefault(
          main?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        ),
        user: firstNonDefault(
          main?.params?.user,
          defaults?.params?.user,
          hard_defaults?.params?.user
        ),
        visibleto: convertOpsGenieTargetFromString(
          firstNonDefault(
            main?.params?.visibleto as string,
            defaults?.params?.visibleto as string,
            hard_defaults?.params?.visibleto as string
          )
        ),
      },
    }),
    [main, defaults, hard_defaults]
  );

  return (
    <>
      <NotifyOptions
        name={name}
        main={main?.options}
        defaults={defaults?.options}
        hard_defaults={hard_defaults?.options}
      />
      <FormLabel text="URL Fields" heading />
      <>
        <FormItem
          name={`${name}.url_fields.host`}
          col_sm={9}
          label="Host"
          tooltip="The OpsGenie API host. Use 'api.eu.opsgenie.com' for EU instances"
          defaultVal={convertedDefaults.url_fields.host}
        />
        <FormItem
          name={`${name}.url_fields.port`}
          col_sm={3}
          label="Port"
          isNumber
          defaultVal={convertedDefaults.url_fields.port}
          position="right"
        />
        <FormItem
          name={`${name}.url_fields.apikey`}
          required
          col_sm={12}
          label="API Key"
          defaultVal={convertedDefaults.url_fields.apiKey}
        />
      </>
      <FormLabel text="Params" heading />
      <>
        <FormList
          name={`${name}.params.actions`}
          label="Actions"
          tooltip="Custom actions that will be available for the alert"
          defaults={convertedDefaults.params.actions}
        />
        <FormItem
          name={`${name}.params.alias`}
          label="Alias"
          tooltip="Client-defined identifier of the alert"
          defaultVal={convertedDefaults.params.alias}
        />
        <FormItem
          name={`${name}.params.description`}
          label="Description"
          tooltip="Description field of the alert"
          defaultVal={convertedDefaults.params.description}
          position="right"
        />
        <FormItem
          name={`${name}.params.note`}
          label="Note"
          tooltip="Additional note that will be added while creating the alert"
          defaultVal={convertedDefaults.params.note}
        />
        <FormKeyValMap
          name={`${name}.params.details`}
          label="Details"
          tooltip="Map of key-val custom props of the alert"
          keyPlaceholder="e.g. X-Authorization"
          valuePlaceholder="e.g. 'Bearer TOKEN'"
          defaults={convertedDefaults.params.details}
        />
        <FormItem
          name={`${name}.params.entity`}
          label="Entity"
          tooltip="Entity field of the alert that is generally used to specify which domain the Source field of the alert"
          defaultVal={convertedDefaults.params.entity}
        />
        <FormItem
          name={`${name}.params.priority`}
          label="Priority"
          tooltip="Priority level of the alert. 1/2/3/4/5"
          isNumber
          defaultVal={convertedDefaults.params.priority}
          position="right"
        />
        <OpsGenieTargets
          name={`${name}.params.responders`}
          label="Responders"
          tooltip="Teams, users, escalations and schedules that the alert will be routed to"
          defaults={convertedDefaults.params.responders}
        />
        <FormItem
          name={`${name}.params.source`}
          label="Source"
          tooltip="Source field of the alert"
          defaultVal={convertedDefaults.params.source}
        />
        <FormItem
          name={`${name}.params.tags`}
          label="Tags"
          tooltip="Tags of the alert"
          defaultVal={convertedDefaults.params.tags}
          position="right"
        />
        <FormItem
          name={`${name}.params.title`}
          label="Title"
          tooltip="Notification title, optionally set by the sender"
          defaultVal={convertedDefaults.params.title}
        />
        <FormItem
          name={`${name}.params.user`}
          label="User"
          tooltip="Display name of the request owner"
          defaultVal={convertedDefaults.params.user}
          position="right"
        />
        <OpsGenieTargets
          name={`${name}.params.visibleto`}
          label="Visible To"
          tooltip="Teams and users that the alert will become visible to without sending any notification"
          defaults={convertedDefaults.params.visibleto}
        />
      </>
    </>
  );
};

export default OPSGENIE;
