import { FormItem, FormLabel } from "components/generic/form";

import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { NotifyZulipType } from "types/config";
import { globalOrDefault } from "components/modals/service-edit/util";

/**
 * Returns the form fields for `Zulip Chat`
 *
 * @param name - The path to this `Zulip Chat` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `Zulip Chat` `Notify`
 */
const ZULIP_CHAT = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyZulipType;
  defaults?: NotifyZulipType;
  hard_defaults?: NotifyZulipType;
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
        name={`${name}.url_fields.botmail`}
        required
        label="Bot Mail"
        tooltip="e.g. something@example.com"
        defaultVal={globalOrDefault(
          main?.url_fields?.botmail,
          defaults?.url_fields?.botmail,
          hard_defaults?.url_fields?.botmail
        )}
      />
      <FormItem
        name={`${name}.url_fields.botkey`}
        required
        label="Bot Key"
        defaultVal={globalOrDefault(
          main?.url_fields?.botkey,
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
          main?.url_fields?.host,
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
          main?.params?.stream,
          defaults?.params?.stream,
          hard_defaults?.params?.stream
        )}
      />
      <FormItem
        name={`${name}.params.topic`}
        label="Topic"
        defaultVal={globalOrDefault(
          main?.params?.topic,
          defaults?.params?.topic,
          hard_defaults?.params?.topic
        )}
        onRight
      />
    </>
  </>
);

export default ZULIP_CHAT;
