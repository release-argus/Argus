import { FormLabel, FormTextArea } from "components/generic/form";

import { NotifyGoogleChatType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { firstNonDefault } from "components/modals/service-edit/util";
import { useMemo } from "react";

/**
 * Returns the form fields for `Google Chat`
 *
 * @param name - The path to this `Google Chat` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `Google Chat` `Notify`
 */
const GOOGLE_CHAT = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyGoogleChatType;
  defaults?: NotifyGoogleChatType;
  hard_defaults?: NotifyGoogleChatType;
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        raw: firstNonDefault(
          main?.url_fields?.raw,
          defaults?.url_fields?.raw,
          hard_defaults?.url_fields?.raw
        ),
      },
    }),
    [main, defaults, hard_defaults]
  );

  return (
    <>
      <NotifyOptions
        name={name}
        main={main?.options}
        defaults={defaults?.options}
        hard_defaults={hard_defaults?.options}
      />
      <FormLabel text="URL Fields" heading />
      <>
        <FormTextArea
          name={`${name}.url_fields.raw`}
          required
          col_sm={12}
          rows={2}
          label="Raw"
          tooltip="e.g. chat.googleapis.com/v1/spaces/foo/messages?key=bar&token=baz"
          defaultVal={convertedDefaults.url_fields.raw}
        />
      </>
    </>
  );
};

export default GOOGLE_CHAT;
