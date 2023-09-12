import { FormItem, FormItemColour, FormLabel } from "components/generic/form";

import { NotifyOptions } from "./shared";
import { NotifyTeamsType } from "types/config";
import { globalOrDefault } from "./util";

const TEAMS = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyTeamsType;
  defaults?: NotifyTeamsType;
  hard_defaults?: NotifyTeamsType;
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
        name={`${name}.url_fields.altid`}
        label="Alt ID"
        defaultVal={globalOrDefault(
          global?.url_fields?.altid,
          defaults?.url_fields?.altid,
          hard_defaults?.url_fields?.altid
        )}
      />
      <FormItem
        name={`${name}.url_fields.tenant`}
        label="Tenant"
        defaultVal={globalOrDefault(
          global?.url_fields?.tenant,
          defaults?.url_fields?.tenant,
          hard_defaults?.url_fields?.tenant
        )}
        onRight
      />
      <FormItem
        name={`${name}.url_fields.group`}
        label="Group"
        defaultVal={globalOrDefault(
          global?.url_fields?.group,
          defaults?.url_fields?.group,
          hard_defaults?.url_fields?.group
        )}
      />
      <FormItem
        name={`${name}.url_fields.groupowner`}
        label="Group Owner"
        defaultVal={globalOrDefault(
          global?.url_fields?.groupowner,
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
          global?.params?.color ||
          defaults?.params?.color ||
          hard_defaults?.params?.color
        }
      />
      <FormItem
        name={`${name}.params.host`}
        col_sm={7}
        label="Host"
        defaultVal={globalOrDefault(
          global?.params?.host,
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
          global?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        )}
      />
    </>
  </>
);

export default TEAMS;
