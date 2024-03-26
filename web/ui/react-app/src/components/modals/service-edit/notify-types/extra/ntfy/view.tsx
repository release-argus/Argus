import { FC, memo } from "react";

import { FormItem } from "components/generic/form";
import { NotifyNtfyAction } from "types/config";

interface Props {
  name: string;
  defaults?: NotifyNtfyAction;
}

/**
 * VIEW renders the form fields for the Ntfy action
 *
 * @param name - The name of the field in the form
 * @param defaults - The default values for the action
 * @returns The form fields for this action
 */
const VIEW: FC<Props> = ({ name, defaults }) => (
  <FormItem
    name={`${name}.url`}
    label="URL"
    required
    col_sm={5}
    defaultVal={defaults?.url}
    placeholder="e.g. 'http://example.com'"
    position="right"
  />
);

export default memo(VIEW);
