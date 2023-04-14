import { FormItem, FormLabel } from "components/generic/form";

import { NotifyOptions } from "./generic";
import { NotifyZulipType } from "types/config";
import { useGlobalOrDefault } from "./util";

const ZULIP_CHAT = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyZulipType;
  defaults?: NotifyZulipType;
  hard_defaults?: NotifyZulipType;
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
        name={`${name}.url_fields.botmail`}
        required
        label="Bot Mail"
        tooltip="e.g. something@example.com"
        placeholder={useGlobalOrDefault(
          global?.url_fields?.botmail,
          defaults?.url_fields?.botmail,
          hard_defaults?.url_fields?.botmail
        )}
      />
      <FormItem
        name={`${name}.url_fields.botkey`}
        required
        label="Bot Key"
        placeholder={useGlobalOrDefault(
          global?.url_fields?.botkey,
          defaults?.url_fields?.botkey,
          hard_defaults?.url_fields?.botkey
        )}
        onRight
      />
      <FormItem
        name={`${name}.url_fields.host`}
        required
        col_sm={12}
        label="Host"
        tooltip="e.g. zulip.example.com"
        placeholder={useGlobalOrDefault(
          global?.url_fields?.host,
          defaults?.url_fields?.host,
          hard_defaults?.url_fields?.host
        )}
      />
    </>
    <>
      <FormLabel text="Params" heading />
      <FormItem
        name={`${name}.params.stream`}
        label="Stream"
        placeholder={useGlobalOrDefault(
          global?.params?.stream,
          defaults?.params?.stream,
          hard_defaults?.params?.stream
        )}
      />
      <FormItem
        name={`${name}.params.topic`}
        label="Topic"
        placeholder={useGlobalOrDefault(
          global?.params?.topic,
          defaults?.params?.topic,
          hard_defaults?.params?.topic
        )}
        onRight
      />
    </>
  </>
);

export default ZULIP_CHAT;
