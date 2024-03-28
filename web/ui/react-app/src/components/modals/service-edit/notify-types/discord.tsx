import {
  FormItem,
  FormItemWithPreview,
  FormLabel,
} from "components/generic/form";

import { BooleanWithDefault } from "components/generic";
import { NotifyDiscordType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { firstNonDefault } from "components/modals/service-edit/notify-types/util";
import { strToBool } from "utils";
import { useMemo } from "react";

/**
 * DISCORD renders the form fields for the Discord Notify
 *
 * @param name - The name of the field in the form
 * @param global - The global values for this Discord Notify
 * @param defaults - The default values for the Discord Notify
 * @param hard_defaults - The hard default values for the Discord Notify
 * @returns The form fields for this Discord Notify
 */
const DISCORD = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyDiscordType;
  defaults?: NotifyDiscordType;
  hard_defaults?: NotifyDiscordType;
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        token: firstNonDefault(
          global?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
        ),
        webhookid: firstNonDefault(
          global?.url_fields?.webhookid,
          defaults?.url_fields?.webhookid,
          hard_defaults?.url_fields?.webhookid
        ),
      },
      // Params
      params: {
        avatar: firstNonDefault(
          global?.params?.avatar,
          defaults?.params?.avatar,
          hard_defaults?.params?.avatar
        ),
        splitlines:
          strToBool(
            firstNonDefault(
              global?.splitlines,
              defaults?.splitlines,
              hard_defaults?.splitlines
            )
          ) ?? true,
        title: firstNonDefault(
          global?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        ),
        username: firstNonDefault(
          global?.params?.username,
          defaults?.params?.username,
          hard_defaults?.params?.username
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
          name={`${name}.url_fields.webhookid`}
          required
          label="WebHook ID"
          tooltip={
            <>
              e.g. https://discord.com/api/webhooks/
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
              e.g. https://discord.com/api/webhooks/webhook_id/
              <span className="bold-underline">token</span>
            </>
          }
          defaultVal={convertedDefaults.url_fields.token}
          position="right"
        />
      </>
      <>
        <FormLabel text="Params" heading />
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
