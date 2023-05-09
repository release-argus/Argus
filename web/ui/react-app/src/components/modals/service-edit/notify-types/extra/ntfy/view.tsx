import { FC, memo } from "react";

import { FormItem } from "components/generic/form";
import { NotifyNtfyAction } from "types/config";

interface Props {
  name: string;
  defaults?: NotifyNtfyAction;
}

const VIEW: FC<Props> = ({ name, defaults }) => (
  <>
    <FormItem
      name={`${name}.url`}
      label="URL"
      required
      col_xs={11}
      col_sm={5}
      defaultVal={defaults?.url}
      placeholder="e.g. 'http://example.com'"
      onRight
    />
  </>
);

export default memo(VIEW);
