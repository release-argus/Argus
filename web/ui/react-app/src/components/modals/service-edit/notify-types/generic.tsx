import {
  FormItem,
  FormKeyValMap,
  FormLabel,
  FormSelect,
} from "components/generic/form";
import { NotifyGenericRequestMethods, NotifyGenericType } from "types/config";
import {
  convertHeadersFromString,
  normaliseForSelect,
} from "components/modals/service-edit/util";

import { BooleanWithDefault } from "components/generic";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { globalOrDefault } from "components/modals/service-edit/notify-types/util";
import { strToBool } from "utils";
import { useMemo } from "react";
import { useWatch } from "react-hook-form";

const GenericRequestMethodOptions: {
  label: NotifyGenericRequestMethods;
  value: NotifyGenericRequestMethods;
}[] = (
  [
    "OPTIONS",
    "GET",
    "HEAD",
    "POST",
    "PUT",
    "DELETE",
    "TRACE",
    "CONNECT",
  ] as const
).map((method) => ({ label: method, value: method }));

const GENERIC = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyGenericType;
  defaults?: NotifyGenericType;
  hard_defaults?: NotifyGenericType;
}) => {
  const selectedTemplate = useWatch({ name: `${name}.params.template` });

  const convertedDefaults = useMemo(
    () => ({
      requestMethod: globalOrDefault(
        global?.params?.requestmethod,
        defaults?.params?.requestmethod,
        hard_defaults?.params?.requestmethod
      ).toLowerCase(),
      // URL Fields
      host: globalOrDefault(
        global?.url_fields?.host,
        defaults?.url_fields?.host,
        hard_defaults?.url_fields?.host
      ),
      path: globalOrDefault(
        global?.url_fields?.path,
        defaults?.url_fields?.path,
        hard_defaults?.url_fields?.path
      ),
      port: globalOrDefault(
        global?.url_fields?.port,
        defaults?.url_fields?.port,
        hard_defaults?.url_fields?.port
      ),
      customHeaders: convertHeadersFromString(
        globalOrDefault(
          global?.url_fields?.custom_headers,
          defaults?.url_fields?.custom_headers,
          hard_defaults?.url_fields?.custom_headers
        )
      ),
      json_payload_vars: convertHeadersFromString(
        globalOrDefault(
          global?.url_fields?.json_payload_vars,
          defaults?.url_fields?.json_payload_vars,
          hard_defaults?.url_fields?.json_payload_vars
        )
      ),
      query_vars: convertHeadersFromString(
        globalOrDefault(
          global?.url_fields?.query_vars,
          defaults?.url_fields?.query_vars,
          hard_defaults?.url_fields?.query_vars
        )
      ),
      // Params
      contentType: globalOrDefault(
        global?.params?.contenttype,
        defaults?.params?.contenttype,
        hard_defaults?.params?.contenttype
      ),
      disableTLS:
        strToBool(
          global?.params?.disabletls ||
            defaults?.params?.disabletls ||
            hard_defaults?.params?.disabletls
        ) ?? true,
      messageKey: globalOrDefault(
        global?.params?.messagekey,
        defaults?.params?.messagekey,
        hard_defaults?.params?.messagekey
      ),
      template: globalOrDefault(
        global?.params?.template,
        defaults?.params?.template,
        hard_defaults?.params?.template
      ),
      title: globalOrDefault(
        global?.params?.title,
        defaults?.params?.title,
        hard_defaults?.params?.title
      ),
      titleKey: globalOrDefault(
        global?.params?.titlekey,
        defaults?.params?.titlekey,
        hard_defaults?.params?.titlekey
      ),
    }),
    [global, defaults, hard_defaults]
  );

  const genericRequestMethodOptions: { label: string; value: string }[] =
    useMemo(() => {
      const defaultRequestMethod = normaliseForSelect(
        GenericRequestMethodOptions,
        convertedDefaults.requestMethod
      );

      if (defaultRequestMethod)
        return [
          { value: "", label: `${defaultRequestMethod.label} (default)` },
          ...genericRequestMethodOptions,
        ];

      return GenericRequestMethodOptions;
    }, [convertedDefaults.requestMethod]);

  return (
    <>
      <NotifyOptions
        name={name}
        global={global?.options}
        defaults={defaults?.options}
        hard_defaults={hard_defaults?.options}
      />
      <>
        <>
          <FormLabel text="URL Fields" heading />
          <FormItem
            name={`${name}.url_fields.host`}
            required
            col_sm={12}
            label="Host"
            tooltip="e.g. gotify.example.com"
            defaultVal={convertedDefaults.host}
          />
          <FormItem
            name={`${name}.url_fields.port`}
            col_sm={4}
            type="number"
            label="Port"
            tooltip="e.g. 443"
            defaultVal={convertedDefaults.port}
          />
          <FormItem
            name={`${name}.url_fields.path`}
            col_sm={8}
            label="Path"
            tooltip={
              <>
                {"e.g. mattermost.example.io/"}
                <span className="bold-underline">path</span>
              </>
            }
            defaultVal={convertedDefaults.path}
            onRight
          />
          <FormKeyValMap
            name={`${name}.url_fields.custom_headers`}
            tooltip="Additional HTTP headers"
            defaults={convertedDefaults.customHeaders}
          />
          {selectedTemplate && (
            <FormKeyValMap
              name={`${name}.url_fields.json_payload_vars`}
              label="JSON Payload vars"
              tooltip="Override 'title' and 'message' with 'titleKey' and 'messageKey' respectively"
              keyPlaceholder="e.g. key"
              valuePlaceholder="e.g. value"
              defaults={convertedDefaults.json_payload_vars}
            />
          )}
          <FormKeyValMap
            name={`${name}.url_fields.query_vars`}
            label="Query vars"
            tooltip="If you need to pass a query variable that is reserved, you can prefix it with an underscore"
            keyPlaceholder="e.g. foo"
            valuePlaceholder="e.g. bar"
            defaults={convertedDefaults.query_vars}
          />
        </>
        <FormLabel text="Params" heading />
        <FormSelect
          name={`${name}.params.requestmethod`}
          col_sm={4}
          label="Request Method"
          tooltip="The HTTP request method"
          options={genericRequestMethodOptions}
          onMiddle
        />
        <FormItem
          name={`${name}.params.contenttype`}
          col_sm={8}
          label="Content Type"
          tooltip="The value of the Content-Type header"
          defaultVal={convertedDefaults.contentType}
        />
        <FormItem
          name={`${name}.params.messagekey`}
          col_sm={4}
          type="text"
          label="Message Key"
          tooltip="The key that will be used for the message value"
          defaultVal={convertedDefaults.messageKey}
        />
        <FormItem
          name={`${name}.params.template`}
          col_sm={8}
          type="text"
          label="Template"
          tooltip="The template used for creating the request payload"
          defaultVal={convertedDefaults.template}
        />
        <FormItem
          name={`${name}.params.titlekey`}
          col_sm={4}
          type="text"
          label="Title Key"
          tooltip="The key that will be used for the title value"
          defaultVal={convertedDefaults.titleKey}
        />
        <FormItem
          name={`${name}.params.title`}
          col_sm={8}
          type="text"
          label="Title"
          tooltip="Text prepended to the message"
          defaultVal={convertedDefaults.title}
        />
        <BooleanWithDefault
          name={`${name}.params.disabletls`}
          label="Disable TLS"
          defaultValue={convertedDefaults.disableTLS}
        />
      </>
    </>
  );
};

export default GENERIC;
