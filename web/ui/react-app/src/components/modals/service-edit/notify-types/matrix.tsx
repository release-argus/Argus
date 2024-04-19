import { FormItem, FormLabel } from "components/generic/form";
import { firstNonDefault, strToBool } from "utils";

import { BooleanWithDefault } from "components/generic";
import { NotifyMatrixType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { useMemo } from "react";

/**
 * Returns the form fields for `Matrix`
 *
 * @param name - The path to this `Matrix` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `Matrix` `Notify`
 */
const MATRIX = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyMatrixType;
  defaults?: NotifyMatrixType;
  hard_defaults?: NotifyMatrixType;
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        host: firstNonDefault(
          main?.url_fields?.host,
          defaults?.url_fields?.host,
          hard_defaults?.url_fields?.host
        ),
        password: firstNonDefault(
          main?.url_fields?.password,
          defaults?.url_fields?.password,
          hard_defaults?.url_fields?.password
        ),
        port: firstNonDefault(
          main?.url_fields?.port,
          defaults?.url_fields?.port,
          hard_defaults?.url_fields?.port
        ),
        username: firstNonDefault(
          main?.url_fields?.username,
          defaults?.url_fields?.username,
          hard_defaults?.url_fields?.username
        ),
      },
      // Params
      params: {
        disabletls:
          strToBool(
            firstNonDefault(
              main?.params?.disabletls,
              defaults?.params?.disabletls,
              hard_defaults?.params?.disabletls
            )
          ) ?? false,
        rooms: firstNonDefault(
          main?.params?.rooms,
          defaults?.params?.rooms,
          hard_defaults?.params?.rooms
        ),
      },
    }),
    [main, defaults, hard_defaults]
  );

  return (
    <>
      <NotifyOptions
        name={name}
        main={main?.options}
        defaults={defaults?.options}
        hard_defaults={hard_defaults?.options}
      />
      <FormLabel text="URL Fields" heading />
      <>
        <FormItem
          name={`${name}.url_fields.host`}
          required
          col_sm={9}
          label="Host"
          tooltip="e.g. smtp.example.com"
          defaultVal={convertedDefaults.url_fields.host}
        />
        <FormItem
          name={`${name}.url_fields.port`}
          col_sm={3}
          label="Port"
          tooltip="e.g. 25/465/587/2525"
          isNumber
          defaultVal={convertedDefaults.url_fields.port}
          position="right"
        />
        <FormItem
          name={`${name}.url_fields.username`}
          label="Username"
          tooltip="e.g. something@example.com"
          defaultVal={convertedDefaults.url_fields.username}
        />
        <FormItem
          name={`${name}.url_fields.password`}
          required
          label="Password"
          defaultVal={convertedDefaults.url_fields.password}
          position="right"
        />
      </>
      <FormLabel text="Params" heading />
      <>
        <FormItem
          name={`${name}.params.rooms`}
          col_sm={12}
          label="Rooms"
          tooltip="e.g. !ROOM_ID,ALIAS"
          defaultVal={convertedDefaults.params.rooms}
        />
        <BooleanWithDefault
          name={`${name}.params.disabletls`}
          label="Disable TLS"
          defaultValue={convertedDefaults.params.disabletls}
        />
      </>
    </>
  );
};

export default MATRIX;
