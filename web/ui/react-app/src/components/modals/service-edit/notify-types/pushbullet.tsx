import { FormItem, FormLabel } from "components/generic/form";

import { NotifyOptions } from "./generic";
import { NotifyPushbulletType } from "types/config";
import { useGlobalOrDefault } from "./util";

const PUSHBULLET = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyPushbulletType;
  defaults?: NotifyPushbulletType;
  hard_defaults?: NotifyPushbulletType;
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
        col_sm={12}
        label="Access Token"
        placeholder={useGlobalOrDefault(
          global?.url_fields?.token,
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
        placeholder={useGlobalOrDefault(
          global?.url_fields?.targets,
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
        placeholder={useGlobalOrDefault(
          global?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        )}
      />
    </>
  </>
);

export default PUSHBULLET;
