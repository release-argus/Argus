import { Button, ButtonGroup, Col, Row } from "react-bootstrap";
import { FC, memo } from "react";
import { faMinus, faPlus, faTrash } from "@fortawesome/free-solid-svg-icons";
import { useFieldArray, useFormContext } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormItem } from "components/generic/form";

interface Props {
  name: string;
  removeMe?: () => void;
}

const Command: FC<Props> = ({ name, removeMe }) => {
  const { control } = useFormContext();
  const { fields, append, remove } = useFieldArray({
    control,
    name: name,
  });

  return (
    <Col xs={12}>
      <Row>
        {fields.map(({ id }, argIndex) => (
          <FormItem
            key={id}
            name={`${name}.${argIndex}.arg`}
            required
            requiredIgnorePlaceholder
            placeholder={
              argIndex === 0
                ? `e.g. "/bin/bash"`
                : argIndex === 1
                ? `e.g. "/opt/script.sh"`
                : `e.g. "-arg${argIndex - 1}"`
            }
            onRight={argIndex % 2 === 1}
          />
        ))}
      </Row>

      {removeMe && (
        <Button className="btn-unchecked float-left" onClick={() => removeMe()}>
          <FontAwesomeIcon icon={faTrash} />
        </Button>
      )}
      <ButtonGroup aria-label="Add/Remove arguments" style={{ float: "right" }}>
        <Button
          className="btn-unchecked mb-3"
          style={{ float: "right" }}
          onClick={() => append({ arg: "" }, { shouldFocus: false })}
        >
          <FontAwesomeIcon icon={faPlus} />
        </Button>
        <Button
          className="btn-unchecked mb-3"
          style={{ float: "right" }}
          onClick={() =>
            fields.length < 2 && removeMe
              ? removeMe()
              : remove(fields.length - 1)
          }
        >
          <FontAwesomeIcon icon={faMinus} />
        </Button>
      </ButtonGroup>
    </Col>
  );
};

export default memo(Command);
