import { FormItem, FormKeyValMap, FormLabel } from "components/generic/form";
import {
  convertHeadersFromString,
  convertOpsGenieTargetFromString,
  globalOrDefault,
} from "components/modals/service-edit/util";
import { useEffect, useMemo } from "react";

import { NotifyOpsGenieType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { OpsGenieTargets } from "components/modals/service-edit/notify-types/extra";
import { useFormContext } from "react-hook-form";

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
  const { getValues, setValue } = useFormContext();

  const convertedDefaults = useMemo(
    () => ({
      details: convertHeadersFromString(
        globalOrDefault(
          main?.params?.details as string,
          defaults?.params?.details as string,
          hard_defaults?.params?.details as string
        )
      ),
      responders: convertOpsGenieTargetFromString(
        globalOrDefault(
          main?.params?.responders as string,
          defaults?.params?.responders as string,
          hard_defaults?.params?.responders as string
        )
      ),
      visibleto: convertOpsGenieTargetFromString(
        globalOrDefault(
          main?.params?.visibleto as string,
          defaults?.params?.visibleto as string,
          hard_defaults?.params?.visibleto as string
        )
      ),
    }),
    [main, defaults, hard_defaults]
  );

  useEffect(() => {
    const details = getValues(`${name}.params.details`);

    if (typeof details === "string")
      setValue(`${name}.params.details`, convertHeadersFromString(details));

    const responders = getValues(`${name}.paramms.responders`);
    if (typeof responders === "string")
      setValue(
        `${name}.params.responders`,
        convertOpsGenieTargetFromString(responders)
      );

    const visibleto = getValues(`${name}.params.visibleto`);
    if (typeof visibleto === "string")
      setValue(
        `${name}.params.visibleto`,
        convertOpsGenieTargetFromString(visibleto)
      );
  }, []);

  return (
    <>
      <NotifyOptions
        name={name}
        main={main?.options}
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
          defaultVal={globalOrDefault(
            main?.url_fields?.host,
            defaults?.url_fields?.host,
            hard_defaults?.url_fields?.host
          )}
        />
        <FormItem
          name={`${name}.url_fields.port`}
          col_sm={3}
          label="Port"
          isNumber
          defaultVal={globalOrDefault(
            main?.url_fields?.port,
            defaults?.url_fields?.port,
            hard_defaults?.url_fields?.port
          )}
          position="right"
        />
        <FormItem
          name={`${name}.url_fields.apikey`}
          required
          col_sm={12}
          label="API Key"
          defaultVal={globalOrDefault(
            main?.url_fields?.apikey,
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
          defaultVal={globalOrDefault(
            main?.params?.actions as string,
            defaults?.params?.actions as string,
            hard_defaults?.params?.actions as string
          )}
        />
        <FormItem
          name={`${name}.params.alias`}
          label="Alias"
          tooltip="Client-defined identifier of the alert"
          defaultVal={globalOrDefault(
            main?.params?.alias,
            defaults?.params?.alias,
            hard_defaults?.params?.alias
          )}
        />
        <FormItem
          name={`${name}.params.description`}
          label="Description"
          tooltip="Description field of the alert"
          defaultVal={globalOrDefault(
            main?.params?.description,
            defaults?.params?.description,
            hard_defaults?.params?.description
          )}
          position="right"
        />
        <FormItem
          name={`${name}.params.note`}
          label="Note"
          tooltip="Additional note that will be added while creating the alert"
          defaultVal={globalOrDefault(
            main?.params?.note,
            defaults?.params?.note,
            hard_defaults?.params?.note
          )}
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
          defaultVal={globalOrDefault(
            main?.params?.entity,
            defaults?.params?.entity,
            hard_defaults?.params?.entity
          )}
        />
        <FormItem
          name={`${name}.params.priority`}
          label="Priority"
          tooltip="Priority level of the alert. 1/2/3/4/5"
          isNumber
          defaultVal={globalOrDefault(
            main?.params?.priority,
            defaults?.params?.priority,
            hard_defaults?.params?.priority
          )}
          position="right"
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
          defaultVal={globalOrDefault(
            main?.params?.source,
            defaults?.params?.source,
            hard_defaults?.params?.source
          )}
        />
        <FormItem
          name={`${name}.params.tags`}
          label="Tags"
          tooltip="Tags of the alert"
          defaultVal={globalOrDefault(
            main?.params?.tags,
            defaults?.params?.tags,
            hard_defaults?.params?.tags
          )}
          position="right"
        />
        <FormItem
          name={`${name}.params.title`}
          label="Title"
          tooltip="Notification title, optionally set by the sender"
          defaultVal={globalOrDefault(
            main?.params?.title,
            defaults?.params?.title,
            hard_defaults?.params?.title
          )}
        />
        <FormItem
          name={`${name}.params.user`}
          label="User"
          tooltip="Display name of the request owner"
          defaultVal={globalOrDefault(
            main?.params?.user,
            defaults?.params?.user,
            hard_defaults?.params?.user
          )}
          position="right"
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
