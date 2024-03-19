import { FC, memo } from "react";
import { FormItem, FormKeyValMap } from "components/generic/form";
import { HeaderType, NotifyNtfyAction } from "types/config";

interface Props {
  name: string;
  defaults?: NotifyNtfyAction;
}

const BROADCAST: FC<Props> = ({ name, defaults }) => (
  <>
    <FormItem
      name={`${name}.intent`}
      label="Intent"
      required
      col_xs={12}
      col_sm={5}
      defaultVal={defaults?.intent}
      placeholder="e.g. 'io.heckel.ntfy.USER_ACTION'"
      onRight
    />
    <FormKeyValMap
      name={`${name}.extras`}
      label="Extras"
      tooltip="Android intent extras"
      defaults={defaults?.extras as HeaderType[] | undefined}
      keyPlaceholder="e.g. 'cmd'"
      valuePlaceholder="e.g. 'pic'"
    />
  </>
);

export default memo(BROADCAST);
