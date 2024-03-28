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
 * Command renders fields of a command with any number of arguments
 *
 * @param name - The name of the field in the form
 * @param removeMe - The function to remove the command
 * @returns A set of form fields for this command
 */
const Command: FC<Props> = ({ name, removeMe }) => {
  const { fields, append, remove } = useFieldArray({
    name: name,
  });
  const addItem = useCallback(() => {
    append({ arg: "" }, { shouldFocus: false });
  }, []);
  // remove the last item if it's not the only one or doesn't match the defaults
  const removeLast = useCallback(() => {
    if (fields.length > 1) return remove(fields.length - 1);
    if (removeMe) return removeMe();
    return undefined;
  }, [fields.length]);

  const placeholder = (index: number) => {
    if (index === 0) return `e.g. "/bin/bash"`;
    if (index === 1) return `e.g. "/opt/script.sh"`;
    return `e.g. "-arg${index - 1}"`;
  };

  return (
    <Col xs={12}>
      <Row>
        {fields.map(({ id }, argIndex) => (
          <FormItem
            key={id}
            name={`${name}.${argIndex}.arg`}
            required
            placeholder={placeholder(argIndex)}
            position={argIndex % 2 === 1 ? "right" : "left"}
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
          onClick={removeLast}
          disabled={fields.length === 0}
        >
          <FontAwesomeIcon icon={faMinus} />
        </Button>
      </ButtonGroup>
    </Col>
  );
};

export default memo(Command);
