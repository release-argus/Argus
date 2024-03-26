import { FormLabel, FormTextArea } from "components/generic/form";

import { NotifyGoogleChatType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { globalOrDefault } from "components/modals/service-edit/notify-types/util";
import { useMemo } from "react";

const GOOGLE_CHAT = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyGoogleChatType;
  defaults?: NotifyGoogleChatType;
  hard_defaults?: NotifyGoogleChatType;
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        raw: globalOrDefault(
          global?.url_fields?.raw,
          defaults?.url_fields?.raw,
          hard_defaults?.url_fields?.raw
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
