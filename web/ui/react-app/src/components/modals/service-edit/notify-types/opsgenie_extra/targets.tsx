import { Button, ButtonGroup, Col, Form, Row } from "react-bootstrap";
import { FC, memo, useCallback } from "react";
import { faMinus, faPlus } from "@fortawesome/free-solid-svg-icons";
import { useFieldArray, useFormContext } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormLabel } from "components/generic/form";
import OpsGenieTarget from "./target";

interface Props {
  name: string;
  label: string;
  tooltip: string;
}

const OpsGenieTargets: FC<Props> = ({ name, label, tooltip }) => {
  const { control } = useFormContext();
  const { fields, append, remove } = useFieldArray({
    control,
    name: name,
  });
  const addItem = useCallback(() => {
    append({ type: "team", sub_type: "id", value: "" }, { shouldFocus: false });
  }, [fields]);
  const removeLast = useCallback(() => {
    remove(fields.length - 1);
  }, [fields]);

  return (
    <Form.Group>
      <FormLabel text={label} tooltip={tooltip} />
      {fields.map(({ id }, index) => (
        <Row key={id}>
          <OpsGenieTarget
            name={`${name}.${index}`}
            removeMe={() => remove(index)}
          />
        </Row>
      ))}
      <Row>
        <Col>
          <ButtonGroup style={{ float: "right" }}>
            <Button
              aria-label={`Add new ${label}`}
              className="btn-unchecked mb-1"
              style={{ float: "right" }}
              onClick={addItem}
            >
              <FontAwesomeIcon icon={faPlus} />
            </Button>
            <Button
              aria-label={`Remove last ${label}`}
              className="btn-unchecked mb-1"
              style={{ float: "left" }}
              onClick={removeLast}
            >
              <FontAwesomeIcon icon={faMinus} />
            </Button>
          </ButtonGroup>
        </Col>
      </Row>
    </Form.Group>
  );
};

export default memo(OpsGenieTargets);
