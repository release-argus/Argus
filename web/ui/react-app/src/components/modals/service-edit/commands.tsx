import { Accordion, Button, FormGroup, Row } from "react-bootstrap";
import { FC, memo, useCallback } from "react";

import Command from "./command";
import { useFieldArray } from "react-hook-form";

interface Props {
  name: string;
}

const EditServiceCommands: FC<Props> = ({ name }) => {
  const { fields, append, remove } = useFieldArray({
    name: name,
  });

  const addItem = useCallback(() => {
    append({ args: [{ arg: "" }] }, { shouldFocus: false });
  }, []);

  return (
    <Accordion>
      <Accordion.Header>Command:</Accordion.Header>
      <Accordion.Body>
        <FormGroup className="mb-2">
          <Row>
            {fields.map(({ id }, index) => (
              <Row key={id}>
                <Command
                  name={`${name}.${index}.args`}
                  removeMe={() => remove(index)}
                />
              </Row>
            ))}
          </Row>
          <Row>
            <Button
              className={fields.length > 0 ? "" : "mt-2"}
              variant="secondary"
              onClick={addItem}
            >
              Add Command
            </Button>
          </Row>
        </FormGroup>
      </Accordion.Body>
    </Accordion>
  );
};

export default memo(EditServiceCommands);
