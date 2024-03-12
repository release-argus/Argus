import { Button, ButtonGroup, Col, FormGroup, Row } from "react-bootstrap";
import { FC, memo, useCallback, useEffect } from "react";
import { FormItem, FormLabel } from "components/generic/form";
import { faMinus, faPlus } from "@fortawesome/free-solid-svg-icons";
import { useFieldArray, useFormContext } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

interface Props {
  name: string;
  label?: string;
  tooltip?: string;
  placeholder?: (index: number) => string;

  defaults?: { [key: string]: string }[];
}

const FormList: FC<Props> = ({
  name,
  label = "List",
  tooltip,
  placeholder,

  defaults,
}) => {
  const { getValues } = useFormContext();
  const { fields, append, remove } = useFieldArray({
    name: name,
  });
  const addItem = useCallback(() => {
    append({ arg: "" }, { shouldFocus: false });
  }, []);
  const removeLast = useCallback(() => {
    remove(fields.length - 1);
  }, [fields.length]);

  // on load, give the defaults if not overridden
  useEffect(() => {
    const useDefaults = defaults?.length && getValues(name).length == 0;
    if (useDefaults) {
      defaults?.forEach(() => {
        append({}, { shouldFocus: false });
      });
    }
  }, []);

  return (
    <FormGroup>
      <Row>
        <Col className="pt-1">
          <FormLabel text={label} tooltip={tooltip} />
        </Col>
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
              disabled={fields.length === 0}
            >
              <FontAwesomeIcon icon={faMinus} />
            </Button>
          </ButtonGroup>
        </Col>
      </Row>
      <Row>
        {fields.map(({ id }, index) => (
          <FormItem
            key={id}
            name={`${name}.${index}.arg`}
            required
            placeholder={placeholder ? placeholder(index) : undefined}
            onRight={index % 2 === 1}
          />
        ))}
      </Row>
    </FormGroup>
  );
};

export default memo(FormList);
