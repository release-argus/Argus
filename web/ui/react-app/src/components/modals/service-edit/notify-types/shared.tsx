import { FormItem, FormLabel, FormTextArea } from "components/generic/form";
import { memo, useMemo } from "react";

import { NotifyOptionsType } from "types/config";
import { globalOrDefault } from "components/modals/service-edit/notify-types/util";

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
      delay: globalOrDefault(
        global?.delay,
        defaults?.delay,
        hard_defaults?.delay
      ),
      max_tries: globalOrDefault(
        global?.max_tries,
        defaults?.max_tries,
        hard_defaults?.max_tries
      ),
      message: globalOrDefault(
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
        onRight
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
