import { FormItem, FormLabel, FormTextArea } from "components/generic/form";
import { memo, useMemo } from "react";

import { NotifyOptionsType } from "types/config";
import { firstNonDefault } from "components/modals/service-edit/notify-types/util";

/**
 * NotifyOptions is the form fields for all Notify Options
 *
 * @param name - The name of the field in the form
 * @param global - The global values for this Notify Options
 * @param defaults - The default values for the Notify Options of this type
 * @param hard_defaults - The hard default values for the Notify Options of this type
 * @returns The form fields for this Notify Options
 */
export const NotifyOptions = ({
  name,

  global,
  defaults,
  hard_defaults,
}: {
  name: string;

  global?: NotifyOptionsType;
  defaults?: NotifyOptionsType;
  hard_defaults?: NotifyOptionsType;
}) => {
  const convertedDefaults = useMemo(
    () => ({
      // Options
      delay: firstNonDefault(
        global?.delay,
        defaults?.delay,
        hard_defaults?.delay
      ),
      max_tries: firstNonDefault(
        global?.max_tries,
        defaults?.max_tries,
        hard_defaults?.max_tries
      ),
      message: firstNonDefault(
        global?.message,
        defaults?.message,
        hard_defaults?.message
      ),
    }),
    [global, defaults, hard_defaults]
  );

  return (
    <>
      <FormLabel text="Options" heading />
      <FormItem
        name={`${name}.options.delay`}
        col_xs={6}
        label="Delay"
        tooltip="e.g. 1h2m3s = 1 hour, 2 minutes and 3 seconds"
        defaultVal={convertedDefaults.delay}
      />
      <FormItem
        name={`${name}.options.max_tries`}
        col_xs={6}
        type="number"
        label="Max tries"
        defaultVal={convertedDefaults.max_tries}
        position="right"
      />
      <FormTextArea
        name={`${name}.options.message`}
        col_sm={12}
        rows={3}
        label="Message"
        defaultVal={convertedDefaults.message}
      />
    </>
  );
};

export default memo(NotifyOptions);
