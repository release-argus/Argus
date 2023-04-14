import { FormItem, FormLabel, FormSelect } from "components/generic/form";

import { BooleanWithDefault } from "components/generic";
import { NotifyOptions } from "./generic";
import { NotifySMTPType } from "types/config";
import { useEffect } from "react";
import { useFormContext } from "react-hook-form";
import { useGlobalOrDefault } from "./util";

export const SMTPAuthOptions = [
  { value: "none", label: "None" },
  { value: "plain", label: "Plain" },
  { value: "crammd5", label: "CRAM-MD5" },
  { value: "unknown", label: "Unknown" },
  { value: "oauth2", label: "OAuth2" },
];

const SMTP = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifySMTPType;
  defaults?: NotifySMTPType;
  hard_defaults?: NotifySMTPType;
}) => {
  const { setValue } = useFormContext();
  const defaultParamsAuth = useGlobalOrDefault(
    global?.params?.auth,
    defaults?.params?.auth,
    hard_defaults?.params?.auth
  );
  useEffect(() => {
    global?.params?.auth && setValue(`${name}.params.auth`, "");
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
          name={`${name}.url_fields.username`}
          label="Username"
          tooltip="e.g. something@example.com"
          placeholder={useGlobalOrDefault(
            global?.url_fields?.username,
            defaults?.url_fields?.username,
            hard_defaults?.url_fields?.username
          )}
        />
        <FormItem
          name={`${name}.url_fields.password`}
          label="Password"
          placeholder={useGlobalOrDefault(
            global?.url_fields?.password,
            defaults?.url_fields?.password,
            hard_defaults?.url_fields?.password
          )}
          onRight
        />
        <FormItem
          name={`${name}.url_fields.host`}
          required
          col_sm={9}
          label="Host"
          tooltip="e.g. smtp.example.com"
          placeholder={useGlobalOrDefault(
            global?.url_fields?.host,
            defaults?.url_fields?.host,
            hard_defaults?.url_fields?.host
          )}
        />
        <FormItem
          name={`${name}.url_fields.port`}
          col_sm={3}
          type="number"
          label="Port"
          tooltip="e.g. 25/465/587/2525"
          placeholder={useGlobalOrDefault(
            global?.url_fields?.port,
            defaults?.url_fields?.port,
            hard_defaults?.url_fields?.port
          )}
          onRight
        />
      </>
      <>
        <FormLabel text="Params" heading />
        <FormItem
          name={`${name}.params.toaddresses`}
          required
          col_sm={12}
          label="To Address(es)"
          tooltip="Email(s) to send to (Comma separated)"
          placeholder={useGlobalOrDefault(
            global?.params?.toaddresses,
            defaults?.params?.toaddresses,
            hard_defaults?.params?.toaddresses
          )}
        />
        <FormItem
          name={`${name}.params.fromaddress`}
          required
          label="From Address"
          tooltip="Email to send from"
          placeholder={useGlobalOrDefault(
            global?.params?.fromaddress,
            defaults?.params?.fromaddress,
            hard_defaults?.params?.fromaddress
          )}
        />
        <FormItem
          name={`${name}.params.fromname`}
          label="From Name"
          tooltip="Name to send as"
          placeholder={useGlobalOrDefault(
            global?.params?.fromname,
            defaults?.params?.fromname,
            hard_defaults?.params?.fromname
          )}
          onRight
        />
        <FormSelect
          name={`${name}.params.auth`}
          col_sm={4}
          label="Auth"
          options={
            defaultParamsAuth
              ? [{ value: "", label: defaultParamsAuth }, ...SMTPAuthOptions]
              : SMTPAuthOptions
          }
        />
        <FormItem
          name={`${name}.params.subject`}
          col_sm={8}
          label="Subject"
          tooltip="Email subject"
          placeholder={useGlobalOrDefault(
            global?.params?.subject,
            defaults?.params?.subject,
            hard_defaults?.params?.subject
          )}
          onRight
        />
        <BooleanWithDefault
          name={`${name}.params.usehtml`}
          label="Use HTML"
          tooltip="Whether 'message' is in HTML"
          defaultValue={
            (global?.params?.usehtml ||
              defaults?.params?.usehtml ||
              hard_defaults?.params?.usehtml) === "true"
          }
        />
        <BooleanWithDefault
          name={`${name}.params.usestarttls`}
          label="Use StartTLS"
          tooltip="Use StartTLS encryption"
          defaultValue={
            (global?.params?.usestarttls ||
              defaults?.params?.usestarttls ||
              hard_defaults?.params?.usestarttls) === "true"
          }
        />
      </>
    </>
  );
};

export default SMTP;
