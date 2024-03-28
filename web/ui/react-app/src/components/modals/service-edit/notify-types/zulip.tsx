import { FormItem, FormLabel } from "components/generic/form";

import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { NotifyZulipType } from "types/config";
import { firstNonDefault } from "components/modals/service-edit/notify-types/util";
import { useMemo } from "react";

/**
 * ZULIP_CHAT renders the form fields for the Zulip Chat Notify
 *
 * @param name - The name of the field in the form
 * @param global - The global values for this Zulip Chat Notify
 * @param defaults - The default values for the Zulip Chat Notify
 * @param hard_defaults - The hard default values for the Zulip Chat Notify
 * @returns The form fields for this Zulip Chat Notify
 */
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
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        botkey: firstNonDefault(
          global?.url_fields?.botkey,
          defaults?.url_fields?.botkey,
          hard_defaults?.url_fields?.botkey
        ),
        botmail: firstNonDefault(
          global?.url_fields?.botmail,
          defaults?.url_fields?.botmail,
          hard_defaults?.url_fields?.botmail
        ),
        host: firstNonDefault(
          global?.url_fields?.host,
          defaults?.url_fields?.host,
          hard_defaults?.url_fields?.host
        ),
      },
      // Params
      params: {
        stream: firstNonDefault(
          global?.params?.stream,
          defaults?.params?.stream,
          hard_defaults?.params?.stream
        ),
        topic: firstNonDefault(
          global?.params?.topic,
          defaults?.params?.topic,
          hard_defaults?.params?.topic
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
          name={`${name}.url_fields.botmail`}
          required
          label="Bot Mail"
          tooltip="e.g. something@example.com"
          defaultVal={convertedDefaults.url_fields.botmail}
        />
        <FormItem
          name={`${name}.url_fields.botkey`}
          required
          label="Bot Key"
          defaultVal={convertedDefaults.url_fields.botkey}
          position="right"
        />
        <FormItem
          name={`${name}.url_fields.host`}
          required
          col_sm={12}
          label="Host"
          tooltip="e.g. zulip.example.com"
          defaultVal={convertedDefaults.url_fields.host}
        />
      </>
      <>
        <FormLabel text="Params" heading />
        <FormItem
          name={`${name}.params.stream`}
          label="Stream"
          defaultVal={convertedDefaults.params.stream}
        />
        <FormItem
          name={`${name}.params.topic`}
          label="Topic"
          defaultVal={convertedDefaults.params.topic}
          position="right"
        />
      </>
    </>
  );
};

export default ZULIP_CHAT;
