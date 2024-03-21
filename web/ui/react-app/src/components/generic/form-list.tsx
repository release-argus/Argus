import { Button, ButtonGroup, Col, FormGroup, Row } from "react-bootstrap";
import { FC, memo, useCallback, useEffect, useMemo } from "react";
import { FormItem, FormLabel } from "components/generic/form";
import { faMinus, faPlus } from "@fortawesome/free-solid-svg-icons";
import { useFieldArray, useFormContext, useWatch } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { diffObjects } from "utils/diff-objects";

interface Props {
  name: string;
  label?: string;
  tooltip?: string;

  defaults?: { [key: string]: string }[];
}

const FormList: FC<Props> = ({
  name,
  label = "List",
  tooltip,

  defaults,
}) => {
  const { setValue, trigger } = useFormContext();
  const { fields, append, remove } = useFieldArray({
    name: name,
  });
  const addItem = useCallback(() => {
    append({ arg: "" }, { shouldFocus: false });
  }, []);

  // keep track of the array values so we can switch defaults when they're unchanged
  const fieldValues = useWatch({ name: name });
  // useDefaults when the fieldValues are undefined or the same as the defaults
  const useDefaults = useMemo(
    () =>
      (defaults && diffObjects(fieldValues ?? fields ?? [], defaults)) ?? false,
    [fieldValues, defaults]
  );
  // trigger validation on change of defaults being used/not
  useEffect(() => {
    trigger(name);

    // Give the defaults back if the field is empty
    if ((fieldValues ?? fields ?? [])?.length === 0)
      defaults?.forEach(() => {
        addItem();
      });
  }, [useDefaults]);

  const placeholder = useCallback(
    (index: number) => (useDefaults && defaults?.[index]?.arg) || "",
    [useDefaults, defaults]
  );

  // on load, ensure we don't have another types actions
  // and give the defaults if not overridden
  useEffect(() => {
    for (const item of fieldValues ?? fields ?? []) {
      if ((item.arg || "") === "") {
        setValue(name, []);
        break;
      }
    }
  }, []);

  // remove the last item if it's not the only one or doesn't match the defaults
  const removeLast = useCallback(() => {
    !(useDefaults && fields.length == 1) && remove(fields.length - 1);
  }, [fields.length, useDefaults]);

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
            defaultVal={placeholder(index)}
            onRight={index % 2 === 1}
          />
        ))}
      </Row>
    </FormGroup>
  );
};

export default memo(FormList);
