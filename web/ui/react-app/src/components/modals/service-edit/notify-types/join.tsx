import {
  FormItem,
  FormItemWithPreview,
  FormLabel,
} from "components/generic/form";

import { NotifyJoinType } from "types/config";
import { NotifyOptions } from "./shared";
import { globalOrDefault } from "./util";

const JOIN = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyJoinType;
  defaults?: NotifyJoinType;
  hard_defaults?: NotifyJoinType;
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
        name={`${name}.url_fields.apikey`}
        required
        col_sm={12}
        label="API Key"
        defaultVal={globalOrDefault(
          global?.url_fields?.apikey,
          defaults?.url_fields?.apikey,
          hard_defaults?.url_fields?.apikey
        )}
      />
    </>
    <>
      <FormLabel text="Params" heading />
      <FormItem
        name={`${name}.params.devices`}
        required
        col_sm={12}
        label="Devices"
        tooltip="e.g. ID1,ID2..."
        defaultVal={globalOrDefault(
          global?.params?.devices,
          defaults?.params?.devices,
          hard_defaults?.params?.devices
        )}
      />
      <FormItemWithPreview
        name={`${name}.params.icon`}
        label="Icon"
        tooltip="URL of icon to use"
        defaultVal={
          global?.params?.icon ||
          defaults?.params?.icon ||
          hard_defaults?.params?.icon
        }
      />
      <FormItem
        name={`${name}.params.title`}
        col_sm={12}
        label="Title"
        tooltip="e.g. 'Release - {{ service_id }}'"
        defaultVal={globalOrDefault(
          global?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        )}
      />
    </>
  </>
);

export default JOIN;
