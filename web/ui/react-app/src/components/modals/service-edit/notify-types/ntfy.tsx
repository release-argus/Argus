import {
  FormItem,
  FormItemWithPreview,
  FormLabel,
  FormSelect,
} from "components/generic/form";
import { useEffect, useMemo } from "react";

import { BooleanWithDefault } from "components/generic";
import { NotifyNtfyType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { NtfyActions } from "components/modals/service-edit/notify-types/extra";
import { globalOrDefault } from "components/modals/service-edit/util";
import { normaliseForSelect } from "components/modals/service-edit/util";
import { strToBool } from "utils";
import { useFormContext } from "react-hook-form";

export const NtfySchemeOptions = [
  { label: "HTTPS", value: "https" },
  { label: "HTTP", value: "http" },
];

export const NtfyPriorityOptions = [
  { label: "Min", value: "min" },
  { label: "Low", value: "low" },
  { label: "Default", value: "default" },
  { label: "High", value: "high" },
  { label: "Max", value: "max" },
];

/**
 * Returns the form fields for `NTFY`
 *
 * @param name - The path to this `NTFY` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `NTFY` `Notify`
 */
const NTFY = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyNtfyType;
  defaults?: NotifyNtfyType;
  hard_defaults?: NotifyNtfyType;
}) => {
  const { getValues, setValue } = useFormContext();

  const defaultParamsScheme = globalOrDefault(
    main?.params?.scheme,
    defaults?.params?.scheme,
    hard_defaults?.params?.scheme
  ).toLowerCase();
  const ntfySchemeOptions = useMemo(() => {
    const defaultScheme = normaliseForSelect(
      NtfySchemeOptions,
      defaultParamsScheme
    );

    if (defaultScheme)
      return [
        { value: "", label: `${defaultScheme.label} (default)` },
        ...NtfySchemeOptions,
      ];

    return NtfySchemeOptions;
  }, [defaultParamsScheme]);

  const defaultParamsPriority = globalOrDefault(
    main?.params?.priority,
    defaults?.params?.priority,
    hard_defaults?.params?.priority
  ).toLowerCase();
  const ntfyPriorityOptions = useMemo(() => {
    const defaultPriority = normaliseForSelect(
      NtfyPriorityOptions,
      defaultParamsPriority
    );

    if (defaultPriority)
      return [
        { value: "", label: `${defaultPriority.label} (default)` },
        ...NtfyPriorityOptions,
      ];

    return NtfyPriorityOptions;
  }, [defaultParamsPriority]);

  useEffect(() => {
    // Normalise selected scheme, or default it
    if (defaultParamsScheme === "")
      setValue(
        `${name}.params.scheme`,
        normaliseForSelect(
          NtfySchemeOptions,
          getValues(`${name}.params.scheme`)
        )?.value || "https"
      );

    // Normalise selected priority, or default it
    if (defaultParamsPriority === "")
      setValue(
        `${name}.params.priority`,
        normaliseForSelect(
          NtfyPriorityOptions,
          getValues(`${name}.params.priority`)
        )?.value || "default"
      );
  }, []);

  return (
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
          name={`${name}.url_fields.host`}
          required
          col_sm={9}
          label="Host"
          defaultVal={globalOrDefault(
            main?.url_fields?.host,
            defaults?.url_fields?.host,
            hard_defaults?.url_fields?.host
          )}
        />
        <FormItem
          name={`${name}.url_fields.port`}
          col_sm={3}
          label="Port"
          type="number"
          defaultVal={globalOrDefault(
            main?.url_fields?.port,
            defaults?.url_fields?.port,
            hard_defaults?.url_fields?.port
          )}
          position="right"
        />
        <FormItem
          name={`${name}.url_fields.username`}
          label="Username"
          defaultVal={globalOrDefault(
            main?.url_fields?.username,
            defaults?.url_fields?.username,
            hard_defaults?.url_fields?.username
          )}
        />
        <FormItem
          name={`${name}.url_fields.password`}
          label="Password"
          defaultVal={globalOrDefault(
            main?.url_fields?.password,
            defaults?.url_fields?.password,
            hard_defaults?.url_fields?.password
          )}
          position="right"
        />
        <FormItem
          name={`${name}.url_fields.topic`}
          required
          col_sm={12}
          label="Topic"
          tooltip="Target topic"
          defaultVal={globalOrDefault(
            main?.url_fields?.topic,
            defaults?.url_fields?.topic,
            hard_defaults?.url_fields?.topic
          )}
        />
      </>
      <>
        <FormLabel text="Params" heading />
        <FormSelect
          name={`${name}.params.scheme`}
          col_sm={3}
          label="Scheme"
          tooltip="Server protocol"
          options={ntfySchemeOptions}
        />
        <FormSelect
          name={`${name}.params.priority`}
          col_sm={3}
          label="Priority"
          options={ntfyPriorityOptions}
          position="middle"
        />
        <FormItem
          name={`${name}.params.tags`}
          label="Tags"
          tooltip="Comma-separated list of tags that may or may not map to emojis"
          defaultVal={globalOrDefault(
            main?.params?.tags,
            defaults?.params?.tags,
            hard_defaults?.params?.tags
          )}
          position="right"
        />
        <FormItem
          name={`${name}.params.attach`}
          col_sm={8}
          label="Attach"
          tooltip="URL of an attachment"
          defaultVal={globalOrDefault(
            main?.params?.attach,
            defaults?.params?.attach,
            hard_defaults?.params?.attach
          )}
        />
        <FormItem
          name={`${name}.params.filename`}
          col_sm={4}
          label="Filename"
          tooltip="File name of the attachment"
          defaultVal={globalOrDefault(
            main?.params?.filename,
            defaults?.params?.filename,
            hard_defaults?.params?.filename
          )}
          position="right"
        />
        <FormItem
          name={`${name}.params.email`}
          label="E-mail"
          tooltip="E-mail address to send to"
          defaultVal={globalOrDefault(
            main?.params?.email,
            defaults?.params?.email,
            hard_defaults?.params?.email
          )}
        />
        <FormItem
          name={`${name}.params.title`}
          label="Title"
          defaultVal={globalOrDefault(
            main?.params?.title,
            defaults?.params?.title,
            hard_defaults?.params?.title
          )}
          position="right"
        />
        <FormItem
          name={`${name}.params.click`}
          col_sm={12}
          label="Click"
          tooltip="URL to open when notification is clicked"
          defaultVal={globalOrDefault(
            main?.params?.click,
            defaults?.params?.click,
            hard_defaults?.params?.click
          )}
        />
        <FormItemWithPreview
          name={`${name}.params.icon`}
          label="Icon"
          tooltip="URL to an icon"
          defaultVal={
            main?.params?.icon ||
            defaults?.params?.icon ||
            hard_defaults?.params?.icon
          }
        />
        <NtfyActions
          name={`${name}.params.actions`}
          label="Actions"
          tooltip="Custom action buttons for notifications"
          defaults={globalOrDefault(
            main?.params?.actions as string | undefined,
            defaults?.params?.actions as string | undefined,
            hard_defaults?.params?.actions as string | undefined
          )}
        />
        <BooleanWithDefault
          name={`${name}.params.cache`}
          label="Cache"
          tooltip="Cache messages"
          defaultValue={
            strToBool(
              main?.params?.cache ||
                defaults?.params?.cache ||
                hard_defaults?.params?.cache
            ) ?? true
          }
        />
        <BooleanWithDefault
          name={`${name}.params.firebase`}
          label="Firebase"
          tooltip="Send to Firebase Cloud Messaging"
          defaultValue={
            strToBool(
              main?.params?.firebase ||
                defaults?.params?.firebase ||
                hard_defaults?.params?.firebase
            ) ?? true
          }
        />
      </>
    </>
  );
};
export default NTFY;
