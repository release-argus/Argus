import { Button, ButtonGroup, Col, Form, Row, Stack } from "react-bootstrap";
import { FC, memo, useCallback } from "react";
import { faMinus, faPlus } from "@fortawesome/free-solid-svg-icons";
import { useFieldArray, useFormContext } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import FormKeyVal from "./key-val";
import { FormLabel } from "components/generic/form";

interface Props {
  name: string;
  label?: string;
  tooltip?: string;
  keyPlaceholder?: string;
  valuePlaceholder?: string;
}

const FormKeyValMap: FC<Props> = ({
  name,
  label = "Headers",
  tooltip,
  keyPlaceholder,
  valuePlaceholder,
}) => {
  const { control } = useFormContext();
  const { fields, append, remove } = useFieldArray({
    control,
    name: name,
  });
  const addItem = useCallback(() => {
    append({ key: "", value: "" }, { shouldFocus: false });
  }, [fields]);
  const removeLast = useCallback(() => {
    remove(fields.length - 1);
  }, [fields]);

  return (
    <Form.Group>
      <FormLabel text={label} tooltip={tooltip} />
      <Stack gap={3}>
        {fields.map(({ id }, index) => (
          <Row key={id}>
            <FormKeyVal
              name={`${name}.${index}`}
              removeMe={() => remove(index)}
              keyPlaceholder={keyPlaceholder}
              valuePlaceholder={valuePlaceholder}
            />
          </Row>
        ))}
      </Stack>
      <Row>
        <Col>
          <ButtonGroup style={{ float: "right" }}>
            <Button
              aria-label={`Add new ${label}`}
              className="btn-unchecked mb-1"
              variant="success"
              style={{ float: "right" }}
              onClick={addItem}
            >
              <FontAwesomeIcon icon={faPlus} />
            </Button>
            <Button
              aria-label={`Remove last ${label}`}
              className="btn-unchecked mb-1"
              variant="danger"
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

export default memo(FormKeyValMap);
