import { FormItem, FormLabel } from "components/generic/form";

import { NotifyOptions } from "./generic";
import { NotifyRocketChatType } from "types/config";
import { globalOrDefault } from "./util";

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
        name={`${name}.url_fields.username`}
        col_sm={12}
        label="Username"
        defaultVal={globalOrDefault(
          global?.url_fields?.username,
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
          global?.url_fields?.host,
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
          global?.url_fields?.port,
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
          global?.url_fields?.path,
          defaults?.url_fields?.path,
          hard_defaults?.url_fields?.path
        )}
      />
      <FormItem
        name={`${name}.url_fields.channel`}
        required
        label="Channel"
        defaultVal={globalOrDefault(
          global?.url_fields?.channel,
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
          global?.url_fields?.tokena,
          defaults?.url_fields?.tokena,
          hard_defaults?.url_fields?.tokena
        )}
      />
      <FormItem
        name={`${name}.url_fields.tokenb`}
        required
        label="Token B"
        defaultVal={globalOrDefault(
          global?.url_fields?.tokenb,
          defaults?.url_fields?.tokenb,
          hard_defaults?.url_fields?.tokenb
        )}
        onRight
      />
    </>
  </>
);

export default ROCKET_CHAT;
