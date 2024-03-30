import { FormItem, FormLabel } from "components/generic/form";

import { NotifyOptions } from "./shared";
import { NotifyRocketChatType } from "types/config";
import { globalOrDefault } from "../util/util";

/**
 * Returns the form fields for `Rocket.Chat`
 *
 * @param name - The path to this `Rocket.Chat` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `Rocket.Chat` `Notify`
 */
const ROCKET_CHAT = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyRocketChatType;
  defaults?: NotifyRocketChatType;
  hard_defaults?: NotifyRocketChatType;
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
        name={`${name}.url_fields.username`}
        col_sm={12}
        label="Username"
        defaultVal={globalOrDefault(
          main?.url_fields?.username,
          defaults?.url_fields?.username,
          hard_defaults?.url_fields?.username
        )}
      />
      <FormItem
        name={`${name}.url_fields.host`}
        required
        col_sm={9}
        label="Host"
        tooltip="e.g. rocketchat.example.io"
        defaultVal={globalOrDefault(
          main?.url_fields?.host,
          defaults?.url_fields?.host,
          hard_defaults?.url_fields?.host
        )}
      />
      <FormItem
        required
        name={`${name}.url_fields.port`}
        col_sm={3}
        type="number"
        label="Port"
        defaultVal={globalOrDefault(
          main?.url_fields?.port,
          defaults?.url_fields?.port,
          hard_defaults?.url_fields?.port
        )}
        onRight
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
        defaultVal={globalOrDefault(
          main?.url_fields?.path,
          defaults?.url_fields?.path,
          hard_defaults?.url_fields?.path
        )}
      />
      <FormItem
        name={`${name}.url_fields.channel`}
        required
        label="Channel"
        defaultVal={globalOrDefault(
          main?.url_fields?.channel,
          defaults?.url_fields?.channel,
          hard_defaults?.url_fields?.channel
        )}
        onRight
      />
      <FormItem
        name={`${name}.url_fields.tokena`}
        required
        label="Token A"
        defaultVal={globalOrDefault(
          main?.url_fields?.tokena,
          defaults?.url_fields?.tokena,
          hard_defaults?.url_fields?.tokena
        )}
      />
      <FormItem
        name={`${name}.url_fields.tokenb`}
        required
        label="Token B"
        defaultVal={globalOrDefault(
          main?.url_fields?.tokenb,
          defaults?.url_fields?.tokenb,
          hard_defaults?.url_fields?.tokenb
        )}
        onRight
      />
    </>
  </>
);

export default ROCKET_CHAT;
