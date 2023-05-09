import { FormItem, FormLabel, FormSelect } from "components/generic/form";
import { useEffect, useMemo } from "react";

import { BooleanWithDefault } from "components/generic";
import { NotifyOptions } from "./generic";
import { NotifySMTPType } from "types/config";
import { globalOrDefault } from "./util";
import { normaliseForSelect } from "../util/normalise-selects";
import { strToBool } from "utils";
import { useFormContext } from "react-hook-form";

export const SMTPAuthOptions = [
  { label: "None", value: "None" },
  { label: "Plain", value: "Plain" },
  { label: "CRAM-MD5", value: "CRAMMD5" },
  { label: "Unknown", value: "Unknown" },
  { label: "OAuth2", value: "OAuth2" },
];
export const SMTPEncryptionOptions = [
  { label: "Auto", value: "Auto" },
  { label: "ExplicitTLS", value: "ExplicitTLS" },
  { label: "ImplicitTLS", value: "ImplicitTLS" },
  { label: "None", value: "None" },
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
  const { getValues, setValue } = useFormContext();

  const defaultParamsAuth = globalOrDefault(
    global?.params?.auth,
    defaults?.params?.auth,
    hard_defaults?.params?.auth
  ).toLowerCase();
  const smtpAuthOptions = useMemo(() => {
    const defaultParamsAuthLabel = normaliseForSelect(
      SMTPAuthOptions,
      defaultParamsAuth
    );

    if (defaultParamsAuthLabel)
      return [
        { value: "", label: `${defaultParamsAuthLabel.label} (default)` },
        ...SMTPAuthOptions,
      ];

    return SMTPAuthOptions;
  }, [defaultParamsAuth]);

  const defaultParamsEncryption = globalOrDefault(
    global?.params?.encryption,
    defaults?.params?.encryption,
    hard_defaults?.params?.encryption
  ).toLowerCase();
  const smtpEncryptionOptions = useMemo(() => {
    const defaultParamsEncryptionLabel = normaliseForSelect(
      SMTPEncryptionOptions,
      defaultParamsEncryption
    );

    if (defaultParamsEncryptionLabel)
      return [
        { value: "", label: `${defaultParamsEncryptionLabel.label} (default)` },
        ...SMTPEncryptionOptions,
      ];

    return SMTPEncryptionOptions;
  }, [defaultParamsAuth]);

  useEffect(() => {
    const currentAuth = getValues(`${name}.params.auth`);
    // Normalise selected auth, or default it
    if (defaultParamsAuth === "")
      setValue(
        `${name}.params.auth`,
        normaliseForSelect(SMTPAuthOptions, currentAuth)?.value || "Unknown"
      );

    // Normalise selected encryption, or default it
    if (defaultParamsEncryption === "")
      setValue(
        `${name}.params.encryption`,
        normaliseForSelect(
          SMTPEncryptionOptions,
          getValues(`${name}.params.encryption`)
        )?.value || "Auto"
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
          tooltip="e.g. something@example.com"
          defaultVal={globalOrDefault(
            global?.url_fields?.username,
            defaults?.url_fields?.username,
            hard_defaults?.url_fields?.username
          )}
        />
        <FormItem
          name={`${name}.url_fields.password`}
          label="Password"
          defaultVal={globalOrDefault(
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
          defaultVal={globalOrDefault(
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
          defaultVal={globalOrDefault(
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
          defaultVal={globalOrDefault(
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
          defaultVal={globalOrDefault(
            global?.params?.fromaddress,
            defaults?.params?.fromaddress,
            hard_defaults?.params?.fromaddress
          )}
        />
        <FormItem
          name={`${name}.params.fromname`}
          label="From Name"
          tooltip="Name to send as"
          defaultVal={globalOrDefault(
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
          options={smtpAuthOptions}
        />
        <FormItem
          name={`${name}.params.subject`}
          col_sm={8}
          label="Subject"
          tooltip="Email subject"
          defaultVal={globalOrDefault(
            global?.params?.subject,
            defaults?.params?.subject,
            hard_defaults?.params?.subject
          )}
          onRight
        />
        <FormSelect
          name={`${name}.params.encryption`}
          col_sm={4}
          label="Encryption"
          tooltip="Encryption method"
          options={smtpEncryptionOptions}
        />
        <FormItem
          name={`${name}.params.clienthost`}
          col_sm={8}
          label="Client Host"
          tooltip={`The client host name sent to the SMTP server during HELLO phase. If set to "auto", it will use the OS hostname`}
          defaultVal={globalOrDefault(
            global?.params?.clienthost,
            defaults?.params?.clienthost,
            hard_defaults?.params?.clienthost
          )}
          onRight
        />
        <BooleanWithDefault
          name={`${name}.params.usehtml`}
          label="Use HTML"
          tooltip="Whether 'message' is in HTML"
          defaultValue={
            strToBool(
              global?.params?.usehtml ||
                defaults?.params?.usehtml ||
                hard_defaults?.params?.usehtml
            ) ?? false
          }
        />
        <BooleanWithDefault
          name={`${name}.params.usestarttls`}
          label="Use StartTLS"
          tooltip="Use StartTLS encryption"
          defaultValue={
            strToBool(
              global?.params?.usestarttls ||
                defaults?.params?.usestarttls ||
                hard_defaults?.params?.usestarttls
            ) ?? true
          }
        />
      </>
    </>
  );
};

export default SMTP;
