import { FormItem, FormLabel, FormSelect } from "components/generic/form";
import { useEffect, useMemo } from "react";

import { BooleanWithDefault } from "components/generic";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { NotifyTelegramType } from "types/config";
import { globalOrDefault } from "components/modals/service-edit/notify-types/util";
import { normaliseForSelect } from "components/modals/service-edit/util";
import { strToBool } from "utils";
import { useFormContext } from "react-hook-form";

export const TelegramParseModeOptions = [
  { label: "None", value: "None" },
  { label: "HTML", value: "HTML" },
  { label: "Markdown", value: "Markdown" },
  { label: "Markdown v2", value: "MarkdownV2" },
];

const TELEGRAM = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyTelegramType;
  defaults?: NotifyTelegramType;
  hard_defaults?: NotifyTelegramType;
}) => {
  const { getValues, setValue } = useFormContext();

  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        token: globalOrDefault(
          global?.url_fields?.token,
          defaults?.url_fields?.token,
          hard_defaults?.url_fields?.token
        ),
      },
      // Params
      params: {
        chats: globalOrDefault(
          global?.params?.chats,
          defaults?.params?.chats,
          hard_defaults?.params?.chats
        ),
        notification:
          strToBool(
            global?.params?.notification ||
              defaults?.params?.notification ||
              hard_defaults?.params?.notification
          ) ?? true,
        parsemode: globalOrDefault(
          global?.params?.parsemode,
          defaults?.params?.parsemode,
          hard_defaults?.params?.parsemode
        ).toLowerCase(),
        preview:
          strToBool(
            global?.params?.preview ||
              defaults?.params?.preview ||
              hard_defaults?.params?.preview
          ) || true,
        title: globalOrDefault(
          global?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        ),
      },
    }),
    [global, defaults, hard_defaults]
  );

  const telegramParseModeOptions = useMemo(() => {
    const defaultParseMode = normaliseForSelect(
      TelegramParseModeOptions,
      convertedDefaults.params.parsemode
    );

    if (defaultParseMode)
      return [
        { value: "", label: `${defaultParseMode.label} (default)` },
        ...TelegramParseModeOptions,
      ];

    return TelegramParseModeOptions;
  }, [convertedDefaults.params.parsemode]);

  useEffect(() => {
    // Normalise selected parsemode, or default it
    if (convertedDefaults.params.parsemode === "")
      setValue(
        `${name}.params.parsemode`,
        normaliseForSelect(getValues(`${name}.params.parsemode`))?.value ||
          "None"
      );
  }, []);

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
          name={`${name}.url_fields.token`}
          required
          col_sm={12}
          label="Token"
          defaultVal={convertedDefaults.url_fields.token}
        />
      </>
      <>
        <FormLabel text="Params" heading />
        <FormItem
          name={`${name}.params.chats`}
          required
          col_sm={8}
          label="Chats"
          tooltip="Chat IDs or Channel names, e.g. -123,@bar"
          defaultVal={convertedDefaults.params.chats}
        />
        <FormSelect
          name={`${name}.params.parsemode`}
          col_sm={4}
          label="Parse Mode"
          options={telegramParseModeOptions}
          onRight
        />
        <FormItem
          name={`${name}.params.title`}
          col_sm={12}
          label="Title"
          defaultVal={convertedDefaults.params.title}
        />
        <BooleanWithDefault
          name={`${name}.params.notification`}
          label="Notification"
          tooltip="Disable for silent messages"
          defaultValue={convertedDefaults.params.notification}
        />
        <BooleanWithDefault
          name={`${name}.params.preview`}
          label="Preview"
          tooltip="Enable web page previews on messages"
          defaultValue={convertedDefaults.params.preview}
        />
      </>
    </>
  );
};

export default TELEGRAM;
