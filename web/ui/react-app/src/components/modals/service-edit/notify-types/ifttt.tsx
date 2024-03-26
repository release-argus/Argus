import { FormItem, FormLabel } from "components/generic/form";

import { NotifyIFTTTType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { globalOrDefault } from "components/modals/service-edit/notify-types/util";
import { useMemo } from "react";

/**
 * IFTTT renders the form fields for the IFTTT Notify
 *
 * @param name - The name of the field in the form
 * @param global - The global values for this IFTTT Notify
 * @param defaults - The default values for the IFTTT Notify
 * @param hard_defaults - The hard default values for the IFTTT Notify
 * @returns The form fields for this IFTTT Notify
 */
const IFTTT = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyIFTTTType;
  defaults?: NotifyIFTTTType;
  hard_defaults?: NotifyIFTTTType;
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // URL Fields
      url_fields: {
        webhookid: globalOrDefault(
          global?.url_fields?.webhookid,
          defaults?.url_fields?.webhookid,
          hard_defaults?.url_fields?.webhookid
        ),
      },
      // Params
      params: {
        events: globalOrDefault(
          global?.params?.events,
          defaults?.params?.events,
          hard_defaults?.params?.events
        ),
        title: globalOrDefault(
          global?.params?.title,
          defaults?.params?.title,
          hard_defaults?.params?.title
        ),
        usemessageasvalue: globalOrDefault(
          global?.params?.usemessageasvalue,
          defaults?.params?.usemessageasvalue,
          hard_defaults?.params?.usemessageasvalue
        ),
        usetitleasvalue: globalOrDefault(
          global?.params?.usetitleasvalue,
          defaults?.params?.usetitleasvalue,
          hard_defaults?.params?.usetitleasvalue
        ),
        value1: globalOrDefault(
          global?.params?.value1,
          defaults?.params?.value1,
          hard_defaults?.params?.value1
        ),
        value2: globalOrDefault(
          global?.params?.value2,
          defaults?.params?.value2,
          hard_defaults?.params?.value2
        ),
        value3: globalOrDefault(
          global?.params?.value3,
          defaults?.params?.value3,
          hard_defaults?.params?.value3
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
          name={`${name}.url_fields.webhookid`}
          required
          col_sm={12}
          label="WebHook ID"
          defaultVal={convertedDefaults.url_fields.webhookid}
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
          defaultVal={convertedDefaults.params.events}
        />
        <FormItem
          name={`${name}.params.title`}
          col_sm={12}
          label="Title"
          tooltip="Optional notification title"
          defaultVal={convertedDefaults.params.title}
        />
        <FormItem
          name={`${name}.params.usemessageasvalue`}
          type="number"
          label="Use Message As Value"
          tooltip="Set the corresponding value field to the message"
          defaultVal={convertedDefaults.params.usemessageasvalue}
        />
        <FormItem
          name={`${name}.params.usetitleasvalue`}
          type="number"
          label="Use Title As Value"
          tooltip="Set the corresponding value field to the title"
          defaultVal={convertedDefaults.params.usetitleasvalue}
          position="right"
        />
        <FormItem
          name={`${name}.params.value1`}
          col_sm={4}
          label="Value1"
          defaultVal={convertedDefaults.params.value1}
        />
        <FormItem
          name={`${name}.params.value2`}
          col_sm={4}
          label="Value2"
          defaultVal={convertedDefaults.params.value2}
          position="middle"
        />
        <FormItem
          name={`${name}.params.value3`}
          col_sm={4}
          label="Value3"
          defaultVal={convertedDefaults.params.value3}
          position="right"
        />
      </>
    </>
  );
};

export default IFTTT;
