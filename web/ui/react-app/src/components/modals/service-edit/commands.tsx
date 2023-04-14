import { Accordion, Button, Form, Row } from "react-bootstrap";
import { FC, memo } from "react";
import { useFieldArray, useFormContext } from "react-hook-form";

import Command from "./command";

interface Props {
  name: string;
}

const EditServiceCommands: FC<Props> = ({ name }) => {
  const { control } = useFormContext();
  const { fields, append, remove } = useFieldArray({
    control,
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
