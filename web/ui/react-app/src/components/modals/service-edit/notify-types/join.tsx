import {
  FormItem,
  FormItemWithPreview,
  FormLabel,
} from "components/generic/form";

import { NotifyJoinType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { globalOrDefault } from "components/modals/service-edit/util";

/**
 * Returns the form fields for `Join`
 *
 * @param name - The path to this `Join` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `Join` `Notify`
 */
const JOIN = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyJoinType;
  defaults?: NotifyJoinType;
  hard_defaults?: NotifyJoinType;
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
        name={`${name}.params.devices`}
        required
        col_sm={12}
        label="Devices"
        tooltip="e.g. ID1,ID2..."
        defaultVal={globalOrDefault(
          main?.params?.devices,
          defaults?.params?.devices,
          hard_defaults?.params?.devices
        )}
      />
      <FormItemWithPreview
        name={`${name}.params.icon`}
        label="Icon"
        tooltip="URL of icon to use"
        defaultVal={
          main?.params?.icon ||
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
          main?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        )}
      />
    </>
  </>
);

export default JOIN;
