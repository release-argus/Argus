import { FormItem, FormLabel } from "components/generic/form";

import { BooleanWithDefault } from "components/generic";
import { NotifyMatrixType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { globalOrDefault } from "components/modals/service-edit/notify-types/util";
import { strToBool } from "utils";

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
}) => (
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
        required
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
        name={`${name}.params.rooms`}
        col_sm={12}
        label="Rooms"
        tooltip="e.g. !ROOM_ID,ALIAS"
        defaultVal={globalOrDefault(
          global?.params?.rooms,
          defaults?.params?.rooms,
          hard_defaults?.params?.rooms
        )}
      />
      <BooleanWithDefault
        name={`${name}.params.disabletls`}
        label="Disable TLS"
        defaultValue={
          strToBool(
            global?.params?.disabletls ||
              defaults?.params?.disabletls ||
              hard_defaults?.params?.disabletls
          ) ?? false
        }
      />
    </>
  </>
);

export default MATRIX;
