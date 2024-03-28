import { FormItem, FormLabel } from "components/generic/form";

import { BooleanWithDefault } from "components/generic";
import { NotifyMatrixType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { firstNonDefault } from "components/modals/service-edit/notify-types/util";
import { strToBool } from "utils";
import { useMemo } from "react";

/**
 * MATRIX renders the form fields for the Matrix Notify
 *
 * @param name - The name of the field in the form
 * @param global - The global values for this Matrix Notify
 * @param defaults - The default values for the Matrix Notify
 * @param hard_defaults - The hard default values for the Matrix Notify
 * @returns The form fields for this Matrix Notify
 */
const MATRIX = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyMatrixType;
  defaults?: NotifyMatrixType;
  hard_defaults?: NotifyMatrixType;
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        host: firstNonDefault(
          global?.url_fields?.host,
          defaults?.url_fields?.host,
          hard_defaults?.url_fields?.host
        ),
        password: firstNonDefault(
          global?.url_fields?.password,
          defaults?.url_fields?.password,
          hard_defaults?.url_fields?.password
        ),
        port: firstNonDefault(
          global?.url_fields?.port,
          defaults?.url_fields?.port,
          hard_defaults?.url_fields?.port
        ),
        username: firstNonDefault(
          global?.url_fields?.username,
          defaults?.url_fields?.username,
          hard_defaults?.url_fields?.username
        ),
      },
      // Params
      params: {
        disabletls:
          strToBool(
            firstNonDefault(
              global?.params?.disabletls,
              defaults?.params?.disabletls,
              hard_defaults?.params?.disabletls
            )
          ) ?? false,
        rooms: firstNonDefault(
          global?.params?.rooms,
          defaults?.params?.rooms,
          hard_defaults?.params?.rooms
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
      <>
        <FormLabel text="Params" heading />
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
