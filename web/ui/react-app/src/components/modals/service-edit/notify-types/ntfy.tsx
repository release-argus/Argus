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
import { globalOrDefault } from "components/modals/service-edit/notify-types/util";
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

const NTFY = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyNtfyType;
  defaults?: NotifyNtfyType;
  hard_defaults?: NotifyNtfyType;
}) => {
  const { getValues, setValue } = useFormContext();

  const convertedDefaults = useMemo(
    () => ({
      priority: globalOrDefault(
        global?.params?.priority,
        defaults?.params?.priority,
        hard_defaults?.params?.priority
      ).toLowerCase(),
      scheme: globalOrDefault(
        global?.params?.scheme,
        defaults?.params?.scheme,
        hard_defaults?.params?.scheme
      ).toLowerCase(),
      username: globalOrDefault(
        global?.url_fields?.username,
        defaults?.url_fields?.username,
        hard_defaults?.url_fields?.username
      ),
      password: globalOrDefault(
        global?.url_fields?.password,
        defaults?.url_fields?.password,
        hard_defaults?.url_fields?.password
      ),
      // URL Fields
      host: globalOrDefault(
        global?.url_fields?.host,
        defaults?.url_fields?.host,
        hard_defaults?.url_fields?.host
      ),
      port: globalOrDefault(
        global?.url_fields?.port,
        defaults?.url_fields?.port,
        hard_defaults?.url_fields?.port
      ),
      topic: globalOrDefault(
        global?.url_fields?.topic,
        defaults?.url_fields?.topic,
        hard_defaults?.url_fields?.topic
      ),
      // Params
      actions: globalOrDefault(
        global?.params?.actions as string | undefined,
        defaults?.params?.actions as string | undefined,
        hard_defaults?.params?.actions as string | undefined
      ),
      attach: globalOrDefault(
        global?.params?.attach,
        defaults?.params?.attach,
        hard_defaults?.params?.attach
      ),
      cache:
        strToBool(
          global?.params?.cache ||
            defaults?.params?.cache ||
            hard_defaults?.params?.cache
        ) ?? true,
      click: globalOrDefault(
        global?.params?.click,
        defaults?.params?.click,
        hard_defaults?.params?.click
      ),
      email: globalOrDefault(
        global?.params?.email,
        defaults?.params?.email,
        hard_defaults?.params?.email
      ),
      filename: globalOrDefault(
        global?.params?.filename,
        defaults?.params?.filename,
        hard_defaults?.params?.filename
      ),
      firebase:
        strToBool(
          global?.params?.firebase ||
            defaults?.params?.firebase ||
            hard_defaults?.params?.firebase
        ) ?? true,
      icon: globalOrDefault(
        global?.params?.icon,
        defaults?.params?.icon,
        hard_defaults?.params?.icon
      ),
      tags: globalOrDefault(
        global?.params?.tags,
        defaults?.params?.tags,
        hard_defaults?.params?.tags
      ),
      title: globalOrDefault(
        global?.params?.title,
        defaults?.params?.title,
        hard_defaults?.params?.title
      ),
    }),
    [global, defaults, hard_defaults]
  );

  const ntfyPriorityOptions = useMemo(() => {
    const defaultPriority = normaliseForSelect(
      NtfyPriorityOptions,
      convertedDefaults.priority
    );

    if (defaultPriority)
      return [
        { value: "", label: `${defaultPriority.label} (default)` },
        ...NtfyPriorityOptions,
      ];

    return NtfyPriorityOptions;
  }, [convertedDefaults.priority]);

  const ntfySchemeOptions = useMemo(() => {
    const defaultScheme = normaliseForSelect(
      NtfySchemeOptions,
      convertedDefaults.scheme
    );

    if (defaultScheme)
      return [
        { value: "", label: `${defaultScheme.label} (default)` },
        ...NtfySchemeOptions,
      ];

    return NtfySchemeOptions;
  }, [convertedDefaults.scheme]);

  useEffect(() => {
    // Normalise selected priority, or default it
    if (convertedDefaults.priority === "")
      setValue(
        `${name}.params.priority`,
        normaliseForSelect(
          NtfyPriorityOptions,
          getValues(`${name}.params.priority`)
        )?.value || "default"
      );

    // Normalise selected scheme, or default it
    if (convertedDefaults.scheme === "")
      setValue(
        `${name}.params.scheme`,
        normaliseForSelect(
          NtfySchemeOptions,
          getValues(`${name}.params.scheme`)
        )?.value || "https"
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
          name={`${name}.url_fields.username`}
          label="Username"
          defaultVal={convertedDefaults.username}
        />
        <FormItem
          name={`${name}.url_fields.password`}
          label="Password"
          defaultVal={convertedDefaults.password}
          onRight
        />
        <FormItem
          name={`${name}.url_fields.host`}
          required
          col_sm={9}
          label="Host"
          defaultVal={convertedDefaults.host}
        />
        <FormItem
          name={`${name}.url_fields.port`}
          col_sm={3}
          label="Port"
          type="number"
          defaultVal={convertedDefaults.port}
          onRight
        />
        <FormItem
          name={`${name}.url_fields.topic`}
          required
          col_sm={12}
          label="Topic"
          tooltip="Target topic"
          defaultVal={convertedDefaults.topic}
          onRight
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
          onMiddle
        />
        <FormItem
          name={`${name}.params.tags`}
          label="Tags"
          tooltip="Comma-separated list of tags that may or may not map to emojis"
          defaultVal={convertedDefaults.tags}
          onRight
        />
        <FormItem
          name={`${name}.params.attach`}
          col_sm={8}
          label="Attach"
          tooltip="URL of an attachment"
          defaultVal={convertedDefaults.attach}
        />
        <FormItem
          name={`${name}.params.filename`}
          col_sm={4}
          label="Filename"
          tooltip="File name of the attachment"
          defaultVal={convertedDefaults.filename}
          onRight
        />
        <FormItem
          name={`${name}.params.email`}
          label="E-mail"
          tooltip="E-mail address to send to"
          defaultVal={convertedDefaults.email}
        />
        <FormItem
          name={`${name}.params.title`}
          label="Title"
          defaultVal={convertedDefaults.title}
          onRight
        />
        <FormItem
          name={`${name}.params.click`}
          col_sm={12}
          label="Click"
          tooltip="URL to open when notification is clicked"
          defaultVal={convertedDefaults.click}
        />
        <FormItemWithPreview
          name={`${name}.params.icon`}
          label="Icon"
          tooltip="URL to an icon"
          defaultVal={convertedDefaults.icon}
        />
        <NtfyActions
          name={`${name}.params.actions`}
          label="Actions"
          tooltip="Custom action buttons for notifications"
          defaults={convertedDefaults.actions}
        />
        <BooleanWithDefault
          name={`${name}.params.cache`}
          label="Cache"
          tooltip="Cache messages"
          defaultValue={convertedDefaults.cache}
        />
        <BooleanWithDefault
          name={`${name}.params.firebase`}
          label="Firebase"
          tooltip="Send to Firebase Cloud Messaging"
          defaultValue={convertedDefaults.firebase}
        />
      </>
    </>
  );
};
export default NTFY;
