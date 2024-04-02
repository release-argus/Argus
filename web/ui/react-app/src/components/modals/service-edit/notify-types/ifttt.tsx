import { FormItem, FormLabel } from "components/generic/form";

import { NotifyIFTTTType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { globalOrDefault } from "components/modals/service-edit/util";

/**
 * Returns the form fields for `IFTTT`
 *
 * @param name - The path to this `IFTTT` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `IFTTT` `Notify`
 */
const IFTTT = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyIFTTTType;
  defaults?: NotifyIFTTTType;
  hard_defaults?: NotifyIFTTTType;
}) => (
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
        name={`${name}.url_fields.webhookid`}
        required
        col_sm={12}
        label="WebHook ID"
        defaultVal={globalOrDefault(
          main?.url_fields?.webhookid,
          defaults?.url_fields?.webhookid,
          hard_defaults?.url_fields?.webhookid
        )}
      />
    </>
    <>
      <FormLabel text="Params" heading />
      <FormItem
        name={`${name}.params.events`}
        required
        col_sm={12}
        label="Events"
        tooltip="e.g. event1,event2..."
        defaultVal={globalOrDefault(
          main?.params?.events,
          defaults?.params?.events,
          hard_defaults?.params?.events
        )}
      />
      <FormItem
        name={`${name}.params.title`}
        col_sm={12}
        label="Title"
        tooltip="Optional notification title"
        defaultVal={globalOrDefault(
          main?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        )}
      />
      <FormItem
        name={`${name}.params.usemessageasvalue`}
        label="Use Message As Value"
        tooltip="Set the corresponding value field to the message"
        isNumber
        defaultVal={globalOrDefault(
          main?.params?.usemessageasvalue,
          defaults?.params?.usemessageasvalue,
          hard_defaults?.params?.usemessageasvalue
        )}
      />
      <FormItem
        name={`${name}.params.usetitleasvalue`}
        label="Use Title As Value"
        tooltip="Set the corresponding value field to the title"
        isNumber
        defaultVal={globalOrDefault(
          main?.params?.usetitleasvalue,
          defaults?.params?.usetitleasvalue,
          hard_defaults?.params?.usetitleasvalue
        )}
        position="right"
      />
      <FormItem
        name={`${name}.params.value1`}
        col_sm={4}
        label="Value1"
        defaultVal={globalOrDefault(
          main?.params?.value1,
          defaults?.params?.value1,
          hard_defaults?.params?.value1
        )}
      />
      <FormItem
        name={`${name}.params.value2`}
        col_sm={4}
        label="Value2"
        defaultVal={globalOrDefault(
          main?.params?.value2,
          defaults?.params?.value2,
          hard_defaults?.params?.value2
        )}
        position="middle"
      />
      <FormItem
        name={`${name}.params.value3`}
        col_sm={4}
        label="Value3"
        defaultVal={globalOrDefault(
          main?.params?.value3,
          defaults?.params?.value3,
          hard_defaults?.params?.value3
        )}
        position="right"
      />
    </>
  </>
);

export default IFTTT;
