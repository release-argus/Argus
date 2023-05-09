import { Button, ButtonGroup, Col, Form, Row, Stack } from "react-bootstrap";
import { FC, memo, useCallback, useEffect, useMemo } from "react";
import { faMinus, faPlus } from "@fortawesome/free-solid-svg-icons";
import { useFieldArray, useFormContext, useWatch } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import FormKeyVal from "./form-key-val";
import { FormLabel } from "components/generic/form";
import { HeaderType } from "types/config";
import { diffObjects } from "utils/diff-objects";

interface Props {
  name: string;
  label?: string;
  tooltip?: string;
  keyPlaceholder?: string;
  valuePlaceholder?: string;

  defaults?: HeaderType[];
}

const FormKeyValMap: FC<Props> = ({
  name,
  label = "Headers",
  tooltip,
  keyPlaceholder,
  valuePlaceholder,

  defaults,
}) => {
  const { trigger } = useFormContext();
  const { fields, append, remove } = useFieldArray({
    name: name,
  });
  const addItem = useCallback(() => {
    append({ key: "", value: "" }, { shouldFocus: false });
  }, []);
  const removeLast = useCallback(() => {
    remove(fields.length - 1);
  }, [fields]);

  // keep track of the array values so we can switch defaults when they're unchanged
  const fieldValues = useWatch({ name: name });
  // useDefaults when the fieldValues are undefined or the same as the defaults
  const useDefaults = useMemo(
    () => diffObjects(fieldValues, defaults),
    [fieldValues, defaults]
  );
  useEffect(() => {
    trigger(name);
  }, [useDefaults]);

  // on load, give the defaults if not overridden
  useEffect(() => {
    if (useDefaults) {
      defaults?.forEach(() => {
        append({}, { shouldFocus: false });
      });
    }
  }, []);

  return (
    <Form.Group>
      <Row>
        <Col>
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
            >
              <FontAwesomeIcon icon={faMinus} />
            </Button>
          </ButtonGroup>
        </Col>
      </Row>
      <Stack>
        {fields.map(({ id }, index) => (
          <Row key={id}>
            <FormKeyVal
              name={`${name}.${index}`}
              defaults={useDefaults ? defaults?.[index] : undefined}
              removeMe={() => remove(index)}
              keyPlaceholder={keyPlaceholder}
              valuePlaceholder={valuePlaceholder}
            />
          </Row>
        ))}
      </Stack>
    </Form.Group>
  );
};

export default memo(FormKeyValMap);
