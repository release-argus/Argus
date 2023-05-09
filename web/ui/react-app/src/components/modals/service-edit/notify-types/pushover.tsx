import { FormItem, FormLabel } from "components/generic/form";

import { NotifyOptions } from "./generic";
import { NotifyPushoverType } from "types/config";
import { globalOrDefault } from "./util";

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
        name={`${name}.url_fields.token`}
        required
        col_sm={6}
        label="API Token/Key"
        tooltip="'Create an Application/API Token' on the Pushover dashboard'"
        defaultVal={globalOrDefault(
          global?.url_fields?.token,
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
          global?.url_fields?.user,
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
          global?.params?.devices,
          defaults?.params?.devices,
          hard_defaults?.params?.devices
        )}
      />
      <FormItem
        name={`${name}.params.title`}
        col_sm={9}
        label="Title"
        defaultVal={globalOrDefault(
          global?.params?.title,
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
          global?.params?.priority,
          defaults?.params?.priority,
          hard_defaults?.params?.priority
        )}
        onRight
      />
    </>
  </>
);

export default PUSHOVER;
