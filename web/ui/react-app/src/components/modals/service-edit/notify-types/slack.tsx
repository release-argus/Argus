import {
  FormItem,
  FormItemColour,
  FormItemWithPreview,
  FormLabel,
} from "components/generic/form";

import { NotifyOptions } from "./shared";
import { NotifySlackType } from "types/config";
import { globalOrDefault } from "./util";

const SLACK = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifySlackType;
  defaults?: NotifySlackType;
  hard_defaults?: NotifySlackType;
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
        name={`${name}.url_fields.token`}
        required
        label="Token"
        tooltip={
          <>
            {"xoxb:"}
            <span className="bold-underline">BOT-OAUTH-TOKEN</span>
            {" or "}
            <span className="bold-underline">WEBHOOK</span>
          </>
        }
        defaultVal={globalOrDefault(
          global?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
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
    </>
    <>
      <FormLabel text="Params" heading />
      <FormItem
        name={`${name}.params.botname`}
        label="Bot Name"
        defaultVal={globalOrDefault(
          global?.params?.botname,
          defaults?.params?.botname,
          hard_defaults?.params?.botname
        )}
      />
      <FormItemColour
        name={`${name}.params.color`}
        label="Color"
        tooltip="Message left-hand border color in hex, e.g. #ffffff"
        defaultVal={
          global?.params?.color ||
          defaults?.params?.color ||
          hard_defaults?.params?.color
        }
        onRight
      />
      <FormItemWithPreview
        name={`${name}.params.icon`}
        label="Icon"
        tooltip="Use emoji or URL as icon (based on presence of http(s):// prefix)"
        defaultVal={
          global?.params?.icon ||
          defaults?.params?.icon ||
          hard_defaults?.params?.icon
        }
      />
      <FormItem
        name={`${name}.params.title`}
        col_sm={12}
        type="text"
        label="Title"
        tooltip="Text prepended to the message"
        defaultVal={globalOrDefault(
          global?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        )}
      />
    </>
  </>
);

export default SLACK;
