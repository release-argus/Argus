import { FormItem, FormLabel, FormTextArea } from "components/generic/form";

import { NotifyOptionsType } from "types/config";
import { globalOrDefault } from "../util/util";
import { memo } from "react";

/**
 * Returns the form fields for the `notify.X.options` section
 *
 * @param name - The path to these `options` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for the `options` section of this `Notify`
 */
export const NotifyOptions = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyOptionsType;
  defaults?: NotifyOptionsType;
  hard_defaults?: NotifyOptionsType;
}) => (
  <>
    <FormLabel text="Options" heading />
    <FormItem
      name={`${name}.options.delay`}
      col_xs={6}
      label="Delay"
      tooltip="e.g. 1h2m3s = 1 hour, 2 minutes and 3 seconds"
      defaultVal={globalOrDefault(
        main?.delay,
        defaults?.delay,
        hard_defaults?.delay
      )}
    />
    <FormItem
      name={`${name}.options.max_tries`}
      col_xs={6}
      type="number"
      label="Max tries"
      defaultVal={globalOrDefault(
        main?.max_tries,
        defaults?.max_tries,
        hard_defaults?.max_tries
      )}
      onRight
    />
    <FormTextArea
      name={`${name}.options.message`}
      col_sm={12}
      rows={3}
      label="Message"
      defaultVal={globalOrDefault(
        main?.message,
        defaults?.message,
        hard_defaults?.message
      )}
    />
  </>
);

export default memo(NotifyOptions);
