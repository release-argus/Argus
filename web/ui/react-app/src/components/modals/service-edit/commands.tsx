import { Accordion, Button, Form, Row } from "react-bootstrap";
import { FC, memo } from "react";

import Command from "./command";
import { useFieldArray } from "react-hook-form";

interface Props {
  name: string;
}

const EditServiceCommands: FC<Props> = ({ name }) => {
  const { fields, append, remove } = useFieldArray({
    name: name,
  });

  return (
    <Accordion>
      <Accordion.Header>Command:</Accordion.Header>
      <Accordion.Body>
        <Form.Group className="mb-2">
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
              onClick={() =>
                append({ args: [{ arg: "" }] }, { shouldFocus: false })
              }
            >
              Add Command
            </Button>
          </Row>
        </Form.Group>
      </Accordion.Body>
    </Accordion>
  );
};

export default memo(EditServiceCommands);
