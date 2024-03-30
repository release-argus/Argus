import { Button, ButtonGroup, Col, Row } from "react-bootstrap";
import { FC, memo, useCallback } from "react";
import { faMinus, faPlus, faTrash } from "@fortawesome/free-solid-svg-icons";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormItem } from "components/generic/form";
import { useFieldArray } from "react-hook-form";

interface Props {
  name: string;
  removeMe?: () => void;
}

/**
 * Returns the form fields for a command
 *
 * @param name - The name of the field in the form
 * @param removeMe - The function to remove the command
 * @returns The form fields for this command with any number of arguments
 */
const Command: FC<Props> = ({ name, removeMe }) => {
  const { fields, append, remove } = useFieldArray({
    name: name,
  });
  const addItem = useCallback(() => {
    append({ arg: "" }, { shouldFocus: false });
  }, []);

  return (
    <Col xs={12}>
      <Row>
        {fields.map(({ id }, argIndex) => (
          <FormItem
            key={id}
            name={`${name}.${argIndex}.arg`}
            required
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
          onClick={addItem}
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
