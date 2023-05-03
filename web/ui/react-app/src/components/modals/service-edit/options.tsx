import { Accordion, Form, Row } from "react-bootstrap";
import { FC, memo } from "react";

import { BooleanWithDefault } from "components/generic";
import { FormItem } from "components/generic/form";
import { ServiceOptionsType } from "types/config";
import { useFormContext } from "react-hook-form";

interface Props {
  defaults?: ServiceOptionsType;
  hard_defaults?: ServiceOptionsType;
}

const EditServiceOptions: FC<Props> = ({ defaults, hard_defaults }) => {
  const { register } = useFormContext();
  return (
    <Accordion>
      <Accordion.Header>Options:</Accordion.Header>
      <Accordion.Body>
        <Form.Check label="Active" {...register("options.active")} />
        <FormItem
          key="interval"
          name="options.interval"
          col_sm={12}
          label="Interval"
          placeholder={defaults?.interval || hard_defaults?.interval}
        />
        <Row>
          <BooleanWithDefault
            name="options.semantic_versioning"
            label="Semantic versioning"
            tooltip="Releases follow 'MAJOR.MINOR.PATCH' versioning"
            defaultValue={
              defaults?.semantic_versioning ||
              hard_defaults?.semantic_versioning
            }
          />
        </Row>
      </Accordion.Body>
    </Accordion>
  );
};

export default memo(EditServiceOptions);
