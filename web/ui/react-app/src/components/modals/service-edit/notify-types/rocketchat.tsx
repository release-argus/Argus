import { FormItem, FormLabel } from "components/generic/form";

import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { NotifyRocketChatType } from "types/config";
import { globalOrDefault } from "components/modals/service-edit/notify-types/util";
import { useMemo } from "react";

/**
 * ROCKET_CHAT renders the form fields for the Rocket.Chat Notify
 *
 * @param name - The name of the field in the form
 * @param global - The global values for this Rocket.Chat Notify
 * @param defaults - The default values for the Rocket.Chat Notify
 * @param hard_defaults - The hard default values for the Rocket.Chat Notify
 * @returns The form fields for this Rocket.Chat Notify
 */
const ROCKET_CHAT = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyRocketChatType;
  defaults?: NotifyRocketChatType;
  hard_defaults?: NotifyRocketChatType;
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        channel: globalOrDefault(
          global?.url_fields?.channel,
          defaults?.url_fields?.channel,
          hard_defaults?.url_fields?.channel
        ),
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
        tokena: globalOrDefault(
          global?.url_fields?.tokena,
          defaults?.url_fields?.tokena,
          hard_defaults?.url_fields?.tokena
        ),
        tokenb: globalOrDefault(
          global?.url_fields?.tokenb,
          defaults?.url_fields?.tokenb,
          hard_defaults?.url_fields?.tokenb
        ),
        username: globalOrDefault(
          global?.url_fields?.username,
          defaults?.url_fields?.username,
          hard_defaults?.url_fields?.username
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
          tooltip="e.g. rocketchat.example.io"
          defaultVal={convertedDefaults.url_fields.host}
        />
        <FormItem
          required
          name={`${name}.url_fields.port`}
          col_sm={3}
          type="number"
          label="Port"
          defaultVal={convertedDefaults.url_fields.port}
          position="right"
        />
        <FormItem
          name={`${name}.url_fields.path`}
          label="Path"
          tooltip={
            <>
              e.g. rocketchat.example.io/
              <span className="bold-underline">path</span>
            </>
          }
          defaultVal={convertedDefaults.url_fields.path}
        />
        <FormItem
          name={`${name}.url_fields.channel`}
          required
          label="Channel"
          defaultVal={convertedDefaults.url_fields.channel}
          position="right"
        />
        <FormItem
          name={`${name}.url_fields.username`}
          col_sm={12}
          label="Username"
          defaultVal={convertedDefaults.url_fields.username}
        />
        <FormItem
          name={`${name}.url_fields.tokena`}
          required
          label="Token A"
          defaultVal={convertedDefaults.url_fields.tokena}
        />
        <FormItem
          name={`${name}.url_fields.tokenb`}
          required
          label="Token B"
          defaultVal={convertedDefaults.url_fields.tokenb}
          position="right"
        />
      </>
    </>
  );
};

export default ROCKET_CHAT;
