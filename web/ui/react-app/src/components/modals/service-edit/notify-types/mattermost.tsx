import {
  FormItem,
  FormItemWithPreview,
  FormLabel,
} from "components/generic/form";

import { NotifyMatterMostType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { globalOrDefault } from "components/modals/service-edit/util";

/**
 * Returns the form fields for `MatterMost`
 *
 * @param name - The path to this `MatterMost` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `MatterMost` `Notify`
 */
const MATTERMOST = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyMatterMostType;
  defaults?: NotifyMatterMostType;
  hard_defaults?: NotifyMatterMostType;
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
        name={`${name}.url_fields.host`}
        required
        col_sm={9}
        label="Host"
        tooltip="e.g. gotify.example.com"
        defaultVal={globalOrDefault(
          main?.url_fields?.host,
          defaults?.url_fields?.host,
          hard_defaults?.url_fields?.host
        )}
      />
      <FormItem
        name={`${name}.url_fields.port`}
        col_sm={3}
        label="Port"
        tooltip="e.g. 443"
        isNumber
        defaultVal={globalOrDefault(
          main?.url_fields?.port,
          defaults?.url_fields?.port,
          hard_defaults?.url_fields?.port
        )}
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
        defaultVal={globalOrDefault(
          main?.url_fields?.path,
          defaults?.url_fields?.path,
          hard_defaults?.url_fields?.path
        )}
      />
      <FormItem
        name={`${name}.url_fields.username`}
        label="Username"
        defaultVal={globalOrDefault(
          main?.url_fields?.username,
          defaults?.url_fields?.username,
          hard_defaults?.url_fields?.username
        )}
        position="right"
      />
      <FormItem
        name={`${name}.url_fields.token`}
        required
        label="Token"
        tooltip="WebHook token"
        defaultVal={globalOrDefault(
          main?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
        )}
      />
      <FormItem
        name={`${name}.url_fields.channel`}
        label="Channel"
        tooltip="e.g. releases"
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
      <FormItemWithPreview
        name={`${name}.params.icon`}
        label="Icon"
        tooltip="URL of icon to use"
        defaultVal={
          main?.params?.icon ||
          defaults?.params?.icon ||
          hard_defaults?.params?.icon
        }
      />
    </>
  </>
);

export default MATTERMOST;
