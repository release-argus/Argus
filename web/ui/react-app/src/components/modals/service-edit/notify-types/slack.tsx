import {
  FormItem,
  FormItemColour,
  FormItemWithPreview,
  FormLabel,
} from "components/generic/form";

import { NotifyOptions } from "./shared";
import { NotifySlackType } from "types/config";
import { globalOrDefault } from "../util/util";

/**
 * Returns the form fields for `Slack`
 *
 * @param name - The path to this `Slack` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `Slack` `Notify`
 */
const SLACK = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifySlackType;
  defaults?: NotifySlackType;
  hard_defaults?: NotifySlackType;
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
          main?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
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
        position="right"
      />
    </>
    <>
      <FormLabel text="Params" heading />
      <FormItem
        name={`${name}.params.botname`}
        label="Bot Name"
        defaultVal={globalOrDefault(
          main?.params?.botname,
          defaults?.params?.botname,
          hard_defaults?.params?.botname
        )}
      />
      <FormItemColour
        name={`${name}.params.color`}
        label="Color"
        tooltip="Message left-hand border color in hex, e.g. #ffffff"
        defaultVal={
          main?.params?.color ||
          defaults?.params?.color ||
          hard_defaults?.params?.color
        }
        position="right"
      />
      <FormItemWithPreview
        name={`${name}.params.icon`}
        label="Icon"
        tooltip="Use emoji or URL as icon (based on presence of http(s):// prefix)"
        defaultVal={
          main?.params?.icon ||
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
          main?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        )}
      />
    </>
  </>
);

export default SLACK;
