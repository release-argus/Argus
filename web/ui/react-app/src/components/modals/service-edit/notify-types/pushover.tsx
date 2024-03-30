import { FormItem, FormLabel } from "components/generic/form";

import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { NotifyPushoverType } from "types/config";
import { globalOrDefault } from "components/modals/service-edit/util";

/**
 * Returns the form fields for `PushOver`
 *
 * @param name - The path to this `PushOver` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `PushOver` `Notify`
 */
const PUSHOVER = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyPushoverType;
  defaults?: NotifyPushoverType;
  hard_defaults?: NotifyPushoverType;
}) => (
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
        name={`${name}.url_fields.token`}
        required
        col_sm={6}
        label="API Token/Key"
        tooltip="'Create an Application/API Token' on the Pushover dashboard'"
        defaultVal={globalOrDefault(
          main?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
        )}
      />
      <FormItem
        name={`${name}.url_fields.user`}
        required
        col_sm={6}
        label="User Key"
        tooltip="Top right of Pushover dashboard"
        defaultVal={globalOrDefault(
          main?.url_fields?.user,
          defaults?.url_fields?.user,
          hard_defaults?.url_fields?.user
        )}
        onRight
      />
    </>
    <>
      <FormLabel text="Params" heading />
      <FormItem
        name={`${name}.params.devices`}
        col_sm={12}
        label="Devices"
        tooltip="e.g. device1,device2... (deviceX=Name column in the 'Your Devices' list)"
        defaultVal={globalOrDefault(
          main?.params?.devices,
          defaults?.params?.devices,
          hard_defaults?.params?.devices
        )}
      />
      <FormItem
        name={`${name}.params.title`}
        col_sm={9}
        label="Title"
        defaultVal={globalOrDefault(
          main?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        )}
      />
      <FormItem
        name={`${name}.params.priority`}
        col_sm={3}
        type="number"
        label="Priority"
        tooltip="Only supply priority values between -1 and 1, since 2 requires additional parameters that are not supported yet"
        defaultVal={globalOrDefault(
          main?.params?.priority,
          defaults?.params?.priority,
          hard_defaults?.params?.priority
        )}
        onRight
      />
    </>
  </>
);

export default PUSHOVER;
