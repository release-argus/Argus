import { FormItem, FormLabel } from "components/generic/form";

import { BooleanWithDefault } from "components/generic";
import { NotifyGotifyType } from "types/config";
import { NotifyOptions } from "./generic";
import { useGlobalOrDefault } from "./util";

const GOTIFY = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyGotifyType;
  defaults?: NotifyGotifyType;
  hard_defaults?: NotifyGotifyType;
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
            e.g. gotify.example.io/
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
        name={`${name}.url_fields.token`}
        required
        label="Token"
        placeholder={useGlobalOrDefault(
          global?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
        )}
        onRight
      />
    </>
    <>
      <FormLabel text="Params" heading />
      <BooleanWithDefault
        name={`${name}.params.disabletls`}
        label="Disable TLS"
        defaultValue={
          (global?.params?.disabletls ||
            defaults?.params?.disabletls ||
            hard_defaults?.params?.disabletls) === "true"
        }
      />
      <FormItem
        name={`${name}.params.priority`}
        col_sm={2}
        type="number"
        label="Priority"
        placeholder={useGlobalOrDefault(
          global?.params?.priority,
          defaults?.params?.priority,
          hard_defaults?.params?.priority
        )}
      />
      <FormItem
        name={`${name}.params.title`}
        col_sm={10}
        label="Title"
        placeholder={useGlobalOrDefault(
          global?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        )}
        onRight
      />
    </>
  </>
);

export default GOTIFY;
