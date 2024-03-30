import { FormLabel, FormTextArea } from "components/generic/form";

import { NotifyGoogleChatType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { globalOrDefault } from "components/modals/service-edit/util";

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
      <FormTextArea
        name={`${name}.url_fields.raw`}
        required
        col_sm={12}
        rows={2}
        label="Raw"
        tooltip="e.g. chat.googleapis.com/v1/spaces/foo/messages?key=bar&token=baz"
        defaultVal={globalOrDefault(
          main?.url_fields?.raw,
          defaults?.url_fields?.raw,
          hard_defaults?.url_fields?.raw
        )}
      />
    </>
  </>
);

export default GOOGLE_CHAT;
