import { FormItem, FormLabel, FormTextArea } from "components/generic/form";

import { NotifyOptionsType } from "types/config";
import { memo } from "react";
import { useGlobalOrDefault } from "./util";

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
}) => (
  <>
    <FormLabel text="Options" heading />
    <FormItem
      name={`${name}.options.delay`}
      col_xs={6}
      label="Delay"
      tooltip="e.g. 1h2m3s = 1 hour, 2 minutes and 3 seconds"
      placeholder={useGlobalOrDefault(
        global?.delay,
        defaults?.delay,
        hard_defaults?.delay
      )}
    />
    <FormItem
      name={`${name}.options.max_tries`}
      col_xs={6}
      type="number"
      label="Max tries"
      placeholder={useGlobalOrDefault(
        global?.max_tries,
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
      placeholder={useGlobalOrDefault(
        global?.message,
        defaults?.message,
        hard_defaults?.message
      )}
    />
  </>
);

export default memo(NotifyOptions);
