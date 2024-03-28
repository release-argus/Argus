import {
  FormItem,
  FormItemColour,
  FormItemWithPreview,
  FormLabel,
} from "components/generic/form";

import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { NotifySlackType } from "types/config";
import { firstNonDefault } from "components/modals/service-edit/notify-types/util";
import { useMemo } from "react";

/**
 * SLACK renders the form fields for the Slack Notify
 *
 * @param name - The name of the field in the form
 * @param global - The global values for this Slack Notify
 * @param defaults - The default values for the Slack Notify
 * @param hard_defaults - The hard default values for the Slack Notify
 * @returns The form fields for this Slack Notify
 */
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
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        channel: firstNonDefault(
          global?.url_fields?.channel,
          defaults?.url_fields?.channel,
          hard_defaults?.url_fields?.channel
        ),
        token: firstNonDefault(
          global?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
        ),
      },
      // Params
      params: {
        botname: firstNonDefault(
          global?.params?.botname,
          defaults?.params?.botname,
          hard_defaults?.params?.botname
        ),
        color: firstNonDefault(
          global?.params?.color,
          defaults?.params?.color,
          hard_defaults?.params?.color
        ),
        icon: firstNonDefault(
          global?.params?.icon,
          defaults?.params?.icon,
          hard_defaults?.params?.icon
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
          defaultVal={convertedDefaults.url_fields.token}
        />
        <FormItem
          name={`${name}.url_fields.channel`}
          required
          label="Channel"
          defaultVal={convertedDefaults.url_fields.channel}
          position="right"
        />
      </>
      <>
        <FormLabel text="Params" heading />
        <FormItem
          name={`${name}.params.botname`}
          label="Bot Name"
          defaultVal={convertedDefaults.params.botname}
        />
        <FormItemColour
          name={`${name}.params.color`}
          label="Color"
          tooltip="Message left-hand border color in hex, e.g. #ffffff"
          defaultVal={convertedDefaults.params.color}
          position="right"
        />
        <FormItemWithPreview
          name={`${name}.params.icon`}
          label="Icon"
          tooltip="Use emoji or URL as icon (based on presence of http(s):// prefix)"
          defaultVal={convertedDefaults.params.icon}
        />
        <FormItem
          name={`${name}.params.title`}
          col_sm={12}
          type="text"
          label="Title"
          tooltip="Text prepended to the message"
          defaultVal={convertedDefaults.params.title}
        />
      </>
    </>
  );
};

export default SLACK;
