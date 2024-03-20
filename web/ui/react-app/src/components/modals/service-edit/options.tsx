import { FC, memo, useMemo } from "react";
import { FormCheck, FormItem } from "components/generic/form";

import { Accordion } from "react-bootstrap";
import { BooleanWithDefault } from "components/generic";
import { ServiceOptionsType } from "types/config";

// import { useFormContext } from "react-hook-form";

interface Props {
  defaults?: ServiceOptionsType;
  hard_defaults?: ServiceOptionsType;
}

const EditServiceOptions: FC<Props> = ({ defaults, hard_defaults }) => {
  // const { register } = useFormContext();
  const convertedDefaults = useMemo(
    () => ({
      interval: defaults?.interval || hard_defaults?.interval,
      semantic_versioning:
        defaults?.semantic_versioning || hard_defaults?.semantic_versioning,
    }),
    [defaults, hard_defaults]
  );
  return (
    <Accordion>
      <Accordion.Header>Options:</Accordion.Header>
      <Accordion.Body>
        <FormCheck
        name="options.active"
        label="Active"
        tooltip="Whether the service is active and checking for updates"
        size="sm"
        />
        <FormItem
          key="interval"
          name="options.interval"
          col_sm={12}
          label="Interval"
          tooltip="How often to check for both latest version and deployed version updates"
          defaultVal={convertedDefaults.interval}
        />
        <BooleanWithDefault
          name="options.semantic_versioning"
          label="Semantic versioning"
          tooltip="Releases follow 'MAJOR.MINOR.PATCH' versioning"
          defaultValue={convertedDefaults.semantic_versioning}
        />
      </Accordion.Body>
    </Accordion>
  );
};

export default memo(EditServiceOptions);
