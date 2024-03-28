import { FormItem, FormLabel } from "components/generic/form";

import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { NotifyPushoverType } from "types/config";
import { firstNonDefault } from "components/modals/service-edit/notify-types/util";
import { useMemo } from "react";

/**
 * PUSHOVER renders the form fields for the Pushover Notify
 *
 * @param name - The name of the field in the form
 * @param global - The global values for this Pushover Notify
 * @param defaults - The default values for the Pushover Notify
 * @param hard_defaults - The hard default values for the Pushover Notify
 * @returns The form fields for this Pushover Notify
 */
const PUSHOVER = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyPushoverType;
  defaults?: NotifyPushoverType;
  hard_defaults?: NotifyPushoverType;
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        token: firstNonDefault(
          global?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
        ),
        user: firstNonDefault(
          global?.url_fields?.user,
          defaults?.url_fields?.user,
          hard_defaults?.url_fields?.user
        ),
      },
      // Params
      params: {
        devices: firstNonDefault(
          global?.params?.devices,
          defaults?.params?.devices,
          hard_defaults?.params?.devices
        ),
        priority: firstNonDefault(
          global?.params?.priority,
          defaults?.params?.priority,
          hard_defaults?.params?.priority
        ),
        title: firstNonDefault(
          global?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        ),
      },
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
          name={`${name}.url_fields.token`}
          required
          col_sm={6}
          label="API Token/Key"
          tooltip="'Create an Application/API Token' on the Pushover dashboard'"
          defaultVal={convertedDefaults.url_fields.token}
        />
        <FormItem
          name={`${name}.url_fields.user`}
          required
          col_sm={6}
          label="User Key"
          tooltip="Top right of Pushover dashboard"
          defaultVal={convertedDefaults.url_fields.user}
          position="right"
        />
      </>
      <>
        <FormLabel text="Params" heading />
        <FormItem
          name={`${name}.params.devices`}
          col_sm={12}
          label="Devices"
          tooltip="e.g. device1,device2... (deviceX=Name column in the 'Your Devices' list)"
          defaultVal={convertedDefaults.params.devices}
        />
        <FormItem
          name={`${name}.params.title`}
          col_sm={9}
          label="Title"
          defaultVal={convertedDefaults.params.title}
        />
        <FormItem
          name={`${name}.params.priority`}
          col_sm={3}
          type="number"
          label="Priority"
          tooltip="Only supply priority values between -1 and 1, since 2 requires additional parameters that are not supported yet"
          defaultVal={convertedDefaults.params.priority}
          position="right"
        />
      </>
    </>
  );
};

export default PUSHOVER;
