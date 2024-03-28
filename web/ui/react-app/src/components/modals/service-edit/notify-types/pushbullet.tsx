import { FormItem, FormLabel } from "components/generic/form";

import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { NotifyPushbulletType } from "types/config";
import { firstNonDefault } from "components/modals/service-edit/notify-types/util";
import { useMemo } from "react";

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
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        targets: firstNonDefault(
          global?.url_fields?.targets,
          defaults?.url_fields?.targets,
          hard_defaults?.url_fields?.targets
        ),
        token: firstNonDefault(
          global?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
        ),
      },
      // Params
      params: {
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
          col_sm={12}
          label="Access Token"
          defaultVal={convertedDefaults.url_fields.token}
        />
        <FormItem
          name={`${name}.url_fields.targets`}
          required
          col_sm={12}
          label="Targets"
          tooltip="e.g. DEVICE1,DEVICE2..."
          defaultVal={convertedDefaults.url_fields.targets}
        />
      </>
      <>
        <FormLabel text="Params" heading />
        <FormItem
          name={`${name}.params.title`}
          col_sm={12}
          label="Title"
          defaultVal={convertedDefaults.params.title}
        />
      </>
    </>
  );
};

export default PUSHBULLET;
