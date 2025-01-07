import {
  FormItem,
  FormItemWithPreview,
  FormLabel,
} from "components/generic/form";
import { firstNonDefault, strToBool } from "utils";

import { BooleanWithDefault } from "components/generic";
import { NotifyDiscordType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { useMemo } from "react";

/**
 * Returns the form fields for `Discord`
 *
 * @param name - The path to this `Discord` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `Discord` `Notify`
 */
const DISCORD = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyDiscordType;
  defaults?: NotifyDiscordType;
  hard_defaults?: NotifyDiscordType;
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        token: firstNonDefault(
          main?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
        ),
        webhookid: firstNonDefault(
          main?.url_fields?.webhookid,
          defaults?.url_fields?.webhookid,
          hard_defaults?.url_fields?.webhookid
        ),
      },
      // Params
      params: {
        avatar: firstNonDefault(
          main?.params?.avatar,
          defaults?.params?.avatar,
          hard_defaults?.params?.avatar
        ),
        splitlines:
          strToBool(
            firstNonDefault(
              main?.splitlines,
              defaults?.splitlines,
              hard_defaults?.splitlines
            )
          ) ?? true,
        title: firstNonDefault(
          main?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        ),
        username: firstNonDefault(
          main?.params?.username,
          defaults?.params?.username,
          hard_defaults?.params?.username
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
          name={`${name}.url_fields.webhookid`}
          required
          label="WebHook ID"
          tooltip={
            <>
              e.g. https://discord.com/api/webhooks/{""}
              <span className="bold-underline">webhook_id</span>
              /token
            </>
          }
          defaultVal={convertedDefaults.url_fields.webhookid}
        />
        <FormItem
          name={`${name}.url_fields.token`}
          required
          label="Token"
          tooltip={
            <>
              e.g. https://discord.com/api/webhooks/webhook_id/{""}
              <span className="bold-underline">token</span>
            </>
          }
          defaultVal={convertedDefaults.url_fields.token}
          position="right"
        />
      </>
      <FormLabel text="Params" heading />
      <>
        <FormItemWithPreview
          name={`${name}.params.avatar`}
          label="Avatar"
          tooltip="Override WebHook avatar with this URL"
          defaultVal={convertedDefaults.params.avatar}
        />
        <FormItem
          name={`${name}.params.username`}
          label="Username"
          tooltip="Override the WebHook username"
          defaultVal={convertedDefaults.params.username}
        />
        <FormItem
          name={`${name}.params.title`}
          label="Title"
          defaultVal={convertedDefaults.params.title}
          position="right"
        />
        <BooleanWithDefault
          name={`${name}.params.splitlines}`}
          label="Split Lines"
          tooltip="Whether to send each line as a separate embedded item"
          defaultValue={convertedDefaults.params.splitlines}
        />
      </>
    </>
  );
};

export default DISCORD;
