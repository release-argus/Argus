import { FormItem, FormLabel } from "components/generic/form";

import { NotifyOptions } from "./shared";
import { NotifyZulipType } from "types/config";
import { globalOrDefault } from "./util";

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
        defaultVal={globalOrDefault(
          global?.url_fields?.botmail,
          defaults?.url_fields?.botmail,
          hard_defaults?.url_fields?.botmail
        )}
      />
      <FormItem
        name={`${name}.url_fields.botkey`}
        required
        label="Bot Key"
        defaultVal={globalOrDefault(
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
        defaultVal={globalOrDefault(
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
        defaultVal={globalOrDefault(
          global?.params?.stream,
          defaults?.params?.stream,
          hard_defaults?.params?.stream
        )}
      />
      <FormItem
        name={`${name}.params.topic`}
        label="Topic"
        defaultVal={globalOrDefault(
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
