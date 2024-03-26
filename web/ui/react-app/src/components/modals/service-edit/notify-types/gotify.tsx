import { FormItem, FormLabel } from "components/generic/form";

import { BooleanWithDefault } from "components/generic";
import { NotifyGotifyType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { globalOrDefault } from "components/modals/service-edit/notify-types/util";
import { strToBool } from "utils";
import { useMemo } from "react";

/**
 * GOTIFY renders the form fields for the Gotify Notify
 *
 * @param name - The name of the field in the form
 * @param global - The global values for this Gotify Notify
 * @param defaults - The default values for the Gotify Notify
 * @param hard_defaults - The hard default values for the Gotify Notify
 * @returns The form fields for this Gotify Notify
 */
const GOTIFY = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyGotifyType;
  defaults?: NotifyGotifyType;
  hard_defaults?: NotifyGotifyType;
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        host: globalOrDefault(
          global?.url_fields?.host,
          defaults?.url_fields?.host,
          hard_defaults?.url_fields?.host
        ),
        path: globalOrDefault(
          global?.url_fields?.path,
          defaults?.url_fields?.path,
          hard_defaults?.url_fields?.path
        ),
        port: globalOrDefault(
          global?.url_fields?.port,
          defaults?.url_fields?.port,
          hard_defaults?.url_fields?.port
        ),
        token: globalOrDefault(
          global?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
        ),
      },
      // Params
      params: {
        disabletls:
          strToBool(
            globalOrDefault(
              global?.params?.disabletls,
              defaults?.params?.disabletls,
              hard_defaults?.params?.disabletls
            )
          ) ?? false,
        priority: globalOrDefault(
          global?.params?.priority,
          defaults?.params?.priority,
          hard_defaults?.params?.priority
        ),
        title: globalOrDefault(
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
          name={`${name}.url_fields.host`}
          required
          col_sm={9}
          label="Host"
          tooltip="e.g. gotify.example.com"
          defaultVal={convertedDefaults.url_fields.host}
        />
        <FormItem
          name={`${name}.url_fields.port`}
          col_sm={3}
          label="Port"
          tooltip="e.g. 443"
          isNumber
          defaultVal={convertedDefaults.url_fields.port}
          position="right"
        />
        <FormItem
          name={`${name}.url_fields.path`}
          label="Path"
          tooltip={
            <>
              e.g. gotify.example.io/
              <span className="bold-underline">path</span>
            </>
          }
          defaultVal={convertedDefaults.url_fields.path}
        />
        <FormItem
          name={`${name}.url_fields.token`}
          required
          label="Token"
          defaultVal={convertedDefaults.url_fields.token}
          position="right"
        />
      </>
      <>
        <FormLabel text="Params" heading />
        <FormItem
          name={`${name}.params.priority`}
          col_sm={2}
          type="number"
          label="Priority"
          defaultVal={convertedDefaults.params.priority}
        />
        <FormItem
          name={`${name}.params.title`}
          col_sm={10}
          label="Title"
          defaultVal={convertedDefaults.params.title}
          position="right"
        />
        <BooleanWithDefault
          name={`${name}.params.disabletls`}
          label="Disable TLS"
          defaultValue={convertedDefaults.params.disabletls}
        />
      </>
    </>
  );
};

export default GOTIFY;
