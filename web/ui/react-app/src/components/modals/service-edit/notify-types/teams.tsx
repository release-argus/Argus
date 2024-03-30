import { FormItem, FormItemColour, FormLabel } from "components/generic/form";

import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { NotifyTeamsType } from "types/config";
import { globalOrDefault } from "components/modals/service-edit/util";

/**
 * Returns the form fields for `Teams`
 *
 * @param name - The path to this `Teams` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `Teams` `Notify`
 */
const TEAMS = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyTeamsType;
  defaults?: NotifyTeamsType;
  hard_defaults?: NotifyTeamsType;
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
        name={`${name}.url_fields.altid`}
        label="Alt ID"
        defaultVal={globalOrDefault(
          main?.url_fields?.altid,
          defaults?.url_fields?.altid,
          hard_defaults?.url_fields?.altid
        )}
      />
      <FormItem
        name={`${name}.url_fields.tenant`}
        label="Tenant"
        defaultVal={globalOrDefault(
          main?.url_fields?.tenant,
          defaults?.url_fields?.tenant,
          hard_defaults?.url_fields?.tenant
        )}
        onRight
      />
      <FormItem
        name={`${name}.url_fields.group`}
        label="Group"
        defaultVal={globalOrDefault(
          main?.url_fields?.group,
          defaults?.url_fields?.group,
          hard_defaults?.url_fields?.group
        )}
      />
      <FormItem
        name={`${name}.url_fields.groupowner`}
        label="Group Owner"
        defaultVal={globalOrDefault(
          main?.url_fields?.groupowner,
          defaults?.url_fields?.groupowner,
          hard_defaults?.url_fields?.groupowner
        )}
        onRight
      />
    </>
    <>
      <FormLabel text="Params" heading />
      <FormItemColour
        name={`${name}.params.color`}
        col_sm={5}
        label="Color"
        defaultVal={
          main?.params?.color ||
          defaults?.params?.color ||
          hard_defaults?.params?.color
        }
      />
      <FormItem
        name={`${name}.params.host`}
        col_sm={7}
        label="Host"
        defaultVal={globalOrDefault(
          main?.params?.host,
          defaults?.params?.host,
          hard_defaults?.params?.host
        )}
        onRight
      />
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

export default TEAMS;
