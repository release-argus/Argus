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
import { useEffect, useMemo } from "react";
import { useFormContext, useWatch } from "react-hook-form";

import { BooleanWithDefault } from "components/generic";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { globalOrDefault } from "components/modals/service-edit/util";
import { strToBool } from "utils";

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

/**
 * Returns the form fields for `Generic`
 *
 * @param name - The path to this `Generic` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `Generic` `Notify`
 */
const GENERIC = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyGenericType;
  defaults?: NotifyGenericType;
  hard_defaults?: NotifyGenericType;
}) => {
  const { getValues, setValue } = useFormContext();

  useEffect(() => {
    const header_url_fields = [
      "custom_headers",
      "json_payload_vars",
      "query_vars",
    ];

    for (const field of header_url_fields) {
      const value = getValues(`${name}.url_fields.${field}`);

      if (typeof value === "string")
        setValue(
          `${name}.url_fields.${field}`,
          convertHeadersFromString(value)
        );
    }
  }, []);

  const selectedTemplate = useWatch({ name: `${name}.params.template` });

  const defaultParamsRequestMethod = globalOrDefault(
    main?.params?.requestmethod,
    defaults?.params?.requestmethod,
    hard_defaults?.params?.requestmethod
  ).toLowerCase();
  const genericRequestMethodOptions: { label: string; value: string }[] =
    useMemo(() => {
      const defaultRequestMethod = normaliseForSelect(
        GenericRequestMethodOptions,
        defaultParamsRequestMethod
      );

      if (defaultRequestMethod)
        return [
          { value: "", label: `${defaultRequestMethod.label} (default)` },
          ...GenericRequestMethodOptions,
        ];

      return GenericRequestMethodOptions;
    }, [defaultParamsRequestMethod]);

  return (
    <>
      <NotifyOptions
        name={name}
        main={main?.options}
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
            defaultVal={globalOrDefault(
              main?.url_fields?.host,
              defaults?.url_fields?.host,
              hard_defaults?.url_fields?.host
            )}
          />
          <FormItem
            name={`${name}.url_fields.port`}
            col_sm={4}
            type="number"
            label="Port"
            tooltip="e.g. 443"
            defaultVal={globalOrDefault(
              main?.url_fields?.port,
              defaults?.url_fields?.port,
              hard_defaults?.url_fields?.port
            )}
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
            defaultVal={globalOrDefault(
              main?.url_fields?.path,
              defaults?.url_fields?.path,
              hard_defaults?.url_fields?.path
            )}
            onRight
          />
          <FormKeyValMap
            name={`${name}.url_fields.custom_headers`}
            tooltip="Additional HTTP headers"
          />
          {selectedTemplate && (
            <FormKeyValMap
              name={`${name}.url_fields.json_payload_vars`}
              label="JSON Payload vars"
              tooltip="Override 'title' and 'message' with 'titleKey' and 'messageKey' respectively"
              keyPlaceholder="e.g. key"
              valuePlaceholder="e.g. value"
            />
          )}
          <FormKeyValMap
            name={`${name}.url_fields.query_vars`}
            label="Query vars"
            tooltip="If you need to pass a query variable that is reserved, you can prefix it with an underscore"
            keyPlaceholder="e.g. foo"
            valuePlaceholder="e.g. bar"
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
          defaultVal={globalOrDefault(
            main?.params?.contenttype,
            defaults?.params?.contenttype,
            hard_defaults?.params?.contenttype
          )}
        />
        <FormItem
          name={`${name}.params.messagekey`}
          col_sm={4}
          type="text"
          label="Message Key"
          tooltip="The key that will be used for the message value"
          defaultVal={globalOrDefault(
            main?.params?.messagekey,
            defaults?.params?.messagekey,
            hard_defaults?.params?.messagekey
          )}
        />
        <FormItem
          name={`${name}.params.template`}
          col_sm={8}
          type="text"
          label="Template"
          tooltip="The template used for creating the request payload"
          defaultVal={globalOrDefault(
            main?.params?.template,
            defaults?.params?.template,
            hard_defaults?.params?.template
          )}
        />
        <FormItem
          name={`${name}.params.titlekey`}
          col_sm={4}
          type="text"
          label="Title Key"
          tooltip="The key that will be used for the title value"
          defaultVal={globalOrDefault(
            main?.params?.titlekey,
            defaults?.params?.titlekey,
            hard_defaults?.params?.titlekey
          )}
        />
        <FormItem
          name={`${name}.params.title`}
          col_sm={8}
          type="text"
          label="Title"
          tooltip="Text prepended to the message"
          defaultVal={globalOrDefault(
            main?.params?.title,
            defaults?.params?.title,
            hard_defaults?.params?.title
          )}
        />
        <BooleanWithDefault
          name={`${name}.params.disabletls`}
          label="Disable TLS"
          defaultValue={
            strToBool(
              main?.params?.disabletls ||
                defaults?.params?.disabletls ||
                hard_defaults?.params?.disabletls
            ) ?? true
          }
        />
      </>
    </>
  );
};

export default GENERIC;
