import { FC, memo } from "react";
import { FormCheck, FormItem } from "components/generic/form";

import { Accordion } from "react-bootstrap";
import { BooleanWithDefault } from "components/generic";
import { ServiceOptionsType } from "types/config";

interface Props {
  defaults?: ServiceOptionsType;
  hard_defaults?: ServiceOptionsType;
}

/**
 * Returns the `options` form fields
 *
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for the `options`
 */
const EditServiceOptions: FC<Props> = ({ defaults, hard_defaults }) => (
  <Accordion>
    <Accordion.Header>Options:</Accordion.Header>
    <Accordion.Body>
      <FormCheck
        name="options.active"
        size="sm"
        label="Active"
        tooltip="Whether the service is active and checking for updates"
      />
      <FormItem
        key="interval"
        name="options.interval"
        col_sm={12}
        label="Interval"
        tooltip="How often to check for both latest version and deployed version updates"
        defaultVal={defaults?.interval || hard_defaults?.interval}
      />
      <BooleanWithDefault
        name="options.semantic_versioning"
        label="Semantic versioning"
        tooltip="Releases follow 'MAJOR.MINOR.PATCH' versioning"
        defaultValue={
          defaults?.semantic_versioning || hard_defaults?.semantic_versioning
        }
      />
    </Accordion.Body>
  </Accordion>
);

export default memo(EditServiceOptions);
