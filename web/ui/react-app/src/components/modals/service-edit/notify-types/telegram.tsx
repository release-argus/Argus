import { FormItem, FormLabel, FormSelect } from "components/generic/form";
import { useEffect, useMemo } from "react";

import { BooleanWithDefault } from "components/generic";
import { NotifyOptions } from "./generic";
import { NotifyTelegramType } from "types/config";
import { useFormContext } from "react-hook-form";
import { useGlobalOrDefault } from "./util";

export const TelegramParseModeOptions = [
  { value: "none", label: "None" },
  { value: "markdown", label: "Markdown" },
  { value: "html", label: "HTML" },
  { value: "markdown_v2", label: "Markdown v2" },
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
  const { setValue } = useFormContext();
  const defaultParamsParseMode = useGlobalOrDefault(
    global?.params?.parsemode,
    defaults?.params?.parsemode,
    hard_defaults?.params?.parsemode
  );
  const telegramParseModeOptions = useMemo(
    () =>
      defaultParamsParseMode
        ? [
            { value: "", label: `${defaultParamsParseMode} (default)` },
            ...TelegramParseModeOptions,
          ]
        : TelegramParseModeOptions,
    [defaultParamsParseMode]
  );
  useEffect(() => {
    global?.params?.parsemode && setValue(`${name}.params.parsemode`, "");
  }, [global]);

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
          placeholder={useGlobalOrDefault(
            global?.url_fields?.token,
            defaults?.url_fields?.token,
            hard_defaults?.url_fields?.token
          )}
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
          placeholder={useGlobalOrDefault(
            global?.params?.chats,
            defaults?.params?.chats,
            hard_defaults?.params?.chats
          )}
        />
        <FormSelect
          name={`${name}.params.parsemode`}
          col_sm={4}
          label="Parse Mode"
          options={telegramParseModeOptions}
          onRight
        />
        <BooleanWithDefault
          name={`${name}.params.notification`}
          label="Notification"
          tooltip="Disable for silent messages"
          defaultValue={
            (global?.params?.notification ||
              defaults?.params?.notification ||
              hard_defaults?.params?.notification) === "true"
          }
        />
        <BooleanWithDefault
          name={`${name}.params.preview`}
          label="Preview"
          tooltip="Enable web page previews on messages"
          defaultValue={
            (global?.params?.preview ||
              defaults?.params?.preview ||
              hard_defaults?.params?.preview) === "true"
          }
        />
      </>
    </>
  );
};

export default TELEGRAM;
