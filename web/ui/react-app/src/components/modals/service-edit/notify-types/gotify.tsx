import { FormItem, FormLabel } from "components/generic/form";
import { firstNonDefault, strToBool } from "utils";

import { BooleanWithDefault } from "components/generic";
import { NotifyGotifyType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { useMemo } from "react";

/**
 * Returns the form fields for `Gotify`
 *
 * @param name - The path to this `Gotify` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `Gotify` `Notify`
 */
const GOTIFY = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyGotifyType;
  defaults?: NotifyGotifyType;
  hard_defaults?: NotifyGotifyType;
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        host: firstNonDefault(
          main?.url_fields?.host,
          defaults?.url_fields?.host,
          hard_defaults?.url_fields?.host
        ),
        path: firstNonDefault(
          main?.url_fields?.path,
          defaults?.url_fields?.path,
          hard_defaults?.url_fields?.path
        ),
        port: firstNonDefault(
          main?.url_fields?.port,
          defaults?.url_fields?.port,
          hard_defaults?.url_fields?.port
        ),
        token: firstNonDefault(
          main?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
        ),
      },
      // Params
      params: {
        disabletls:
          strToBool(
            firstNonDefault(
              main?.params?.disabletls,
              defaults?.params?.disabletls,
              hard_defaults?.params?.disabletls
            )
          ) ?? false,
        priority: firstNonDefault(
          main?.params?.priority,
          defaults?.params?.priority,
          hard_defaults?.params?.priority
        ),
        title: firstNonDefault(
          main?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        ),
      },
    }),
    [main, defaults, hard_defaults]
  );

  return (
    <>
      <NotifyOptions
        name={name}
        main={main?.options}
        defaults={defaults?.options}
        hard_defaults={hard_defaults?.options}
      />
      <FormLabel text="URL Fields" heading />
      <>
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
              e.g. gotify.example.io/{""}
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
      <FormLabel text="Params" heading />
      <>
        <FormItem
          name={`${name}.params.priority`}
          col_sm={2}
          label="Priority"
          isNumber
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
