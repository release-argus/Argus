import {
  FormItem,
  FormItemWithPreview,
  FormLabel,
} from "components/generic/form";

import { NotifyMatterMostType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { globalOrDefault } from "components/modals/service-edit/notify-types/util";
import { useMemo } from "react";

/**
 * MATTERMOST renders the form fields for the MatterMost Notify
 *
 * @param name - The name of the field in the form
 * @param global - The global values for this MatterMost Notify
 * @param defaults - The default values for the MatterMost Notify
 * @param hard_defaults - The hard default values for the MatterMost Notify
 * @returns The form fields for this MatterMost Notify
 */
const MATTERMOST = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyMatterMostType;
  defaults?: NotifyMatterMostType;
  hard_defaults?: NotifyMatterMostType;
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
        token: globalOrDefault(
          global?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
        ),
        username: globalOrDefault(
          global?.url_fields?.username,
          defaults?.url_fields?.username,
          hard_defaults?.url_fields?.username
        ),
      },
      // Params
      params: {
        icon: globalOrDefault(
          global?.params?.icon,
          defaults?.params?.icon,
          hard_defaults?.params?.icon
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
          type="number"
          label="Port"
          tooltip="e.g. 443"
          defaultVal={convertedDefaults.url_fields.port}
          position="right"
        />
        <FormItem
          name={`${name}.url_fields.path`}
          label="Path"
          tooltip={
            <>
              {"e.g. mattermost.example.io/"}
              <span className="bold-underline">path</span>
            </>
          }
          defaultVal={convertedDefaults.url_fields.path}
        />
        <FormItem
          name={`${name}.url_fields.username`}
          label="Username"
          defaultVal={convertedDefaults.url_fields.username}
          position="right"
        />
        <FormItem
          name={`${name}.url_fields.token`}
          required
          label="Token"
          tooltip="WebHook token"
          defaultVal={convertedDefaults.url_fields.token}
        />
        <FormItem
          name={`${name}.url_fields.channel`}
          label="Channel"
          tooltip="e.g. releases"
          defaultVal={convertedDefaults.url_fields.channel}
          position="right"
        />
      </>
      <>
        <FormLabel text="Params" heading />
        <FormItemWithPreview
          name={`${name}.params.icon`}
          label="Icon"
          tooltip="URL of icon to use"
          defaultVal={convertedDefaults.params.icon}
        />
      </>
    </>
  );
};

export default MATTERMOST;
