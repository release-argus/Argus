import {
  FormItem,
  FormItemWithPreview,
  FormLabel,
} from "components/generic/form";

import { NotifyMatterMostType } from "types/config";
import { NotifyOptions } from "./generic";
import { useGlobalOrDefault } from "./util";

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
        name={`${name}.url_fields.host`}
        required
        col_sm={9}
        label="Host"
        tooltip="e.g. gotify.example.com"
        placeholder={useGlobalOrDefault(
          global?.url_fields?.host,
          defaults?.url_fields?.host,
          hard_defaults?.url_fields?.host
        )}
      />
      <FormItem
        name={`${name}.url_fields.port`}
        col_sm={3}
        type="number"
        label="Port"
        tooltip="e.g. 443"
        placeholder={useGlobalOrDefault(
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
            {"e.g. mattermost.example.io/"}
            <span className="bold-underline">path</span>
          </>
        }
        placeholder={useGlobalOrDefault(
          global?.url_fields?.path,
          defaults?.url_fields?.path,
          hard_defaults?.url_fields?.path
        )}
      />
      <FormItem
        name={`${name}.url_fields.username`}
        label="Username"
        placeholder={useGlobalOrDefault(
          global?.url_fields?.username,
          defaults?.url_fields?.username,
          hard_defaults?.url_fields?.username
        )}
        onRight
      />
      <FormItem
        name={`${name}.url_fields.token`}
        required
        label="Token"
        tooltip="WebHook token"
        placeholder={useGlobalOrDefault(
          global?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
        )}
      />
      <FormItem
        name={`${name}.url_fields.channel`}
        label="Channel"
        tooltip="e.g. releases"
        placeholder={useGlobalOrDefault(
          global?.url_fields?.channel,
          defaults?.url_fields?.channel,
          hard_defaults?.url_fields?.channel
        )}
        onRight
      />
    </>
    <>
      <FormLabel text="Params" heading />
      <FormItemWithPreview
        name={`${name}.params.icon`}
        label="Icon"
        tooltip="URL of icon to use"
        placeholder={
          global?.params?.icon ||
          defaults?.params?.icon ||
          hard_defaults?.params?.icon
        }
      />
    </>
  </>
);

export default MATTERMOST;
