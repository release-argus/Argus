import { FormItem, FormLabel } from "components/generic/form";

import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { NotifyPushbulletType } from "types/config";
import { globalOrDefault } from "components/modals/service-edit/util";

/**
 * Returns the form fields for `PushBullet`
 *
 * @param name - The path to this `PushBullet` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `PushBullet` `Notify`
 */
const PUSHBULLET = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyPushbulletType;
  defaults?: NotifyPushbulletType;
  hard_defaults?: NotifyPushbulletType;
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
        col_sm={12}
        label="Access Token"
        defaultVal={globalOrDefault(
          main?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
        )}
      />
      <FormItem
        name={`${name}.url_fields.targets`}
        required
        col_sm={12}
        label="Targets"
        tooltip="e.g. DEVICE1,DEVICE2..."
        defaultVal={globalOrDefault(
          main?.url_fields?.targets,
          defaults?.url_fields?.targets,
          hard_defaults?.url_fields?.targets
        )}
      />
    </>
    <>
      <FormLabel text="Params" heading />
      <FormItem
        name={`${name}.params.title`}
        col_sm={12}
        label="Title"
        defaultVal={globalOrDefault(
          main?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        )}
      />
    </>
  </>
);

export default PUSHBULLET;
