import { FormItem, FormItemColour, FormLabel } from "components/generic/form";

import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { NotifyTeamsType } from "types/config";
import { firstNonDefault } from "components/modals/service-edit/notify-types/util";
import { useMemo } from "react";

/**
 * TEAMS renders the form fields for the Teams Notify
 *
 * @param name - The name of the field in the form
 * @param global - The global values for this Teams Notify
 * @param defaults - The default values for the Teams Notify
 * @param hard_defaults - The hard default values for the Teams Notify
 * @returns The form fields for this Teams Notify
 */
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
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        altid: firstNonDefault(
          global?.url_fields?.altid,
          defaults?.url_fields?.altid,
          hard_defaults?.url_fields?.altid
        ),
        group: firstNonDefault(
          global?.url_fields?.group,
          defaults?.url_fields?.group,
          hard_defaults?.url_fields?.group
        ),
        groupowner: firstNonDefault(
          global?.url_fields?.groupowner,
          defaults?.url_fields?.groupowner,
          hard_defaults?.url_fields?.groupowner
        ),
        tenant: firstNonDefault(
          global?.url_fields?.tenant,
          defaults?.url_fields?.tenant,
          hard_defaults?.url_fields?.tenant
        ),
      },
      // Params
      params: {
        color: firstNonDefault(
          global?.params?.color,
          defaults?.params?.color,
          hard_defaults?.params?.color
        ),
        host: firstNonDefault(
          global?.params?.host,
          defaults?.params?.host,
          hard_defaults?.params?.host
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
          name={`${name}.url_fields.altid`}
          label="Alt ID"
          defaultVal={convertedDefaults.url_fields.altid}
        />
        <FormItem
          name={`${name}.url_fields.tenant`}
          label="Tenant"
          defaultVal={convertedDefaults.url_fields.tenant}
          position="right"
        />
        <FormItem
          name={`${name}.url_fields.group`}
          label="Group"
          defaultVal={convertedDefaults.url_fields.group}
        />
        <FormItem
          name={`${name}.url_fields.groupowner`}
          label="Group Owner"
          defaultVal={convertedDefaults.url_fields.groupowner}
          position="right"
        />
      </>
      <>
        <FormLabel text="Params" heading />
        <FormItemColour
          name={`${name}.params.color`}
          col_sm={5}
          label="Color"
          defaultVal={convertedDefaults.params.color}
        />
        <FormItem
          name={`${name}.params.host`}
          col_sm={7}
          label="Host"
          defaultVal={convertedDefaults.params.host}
          position="right"
        />
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

export default TEAMS;
