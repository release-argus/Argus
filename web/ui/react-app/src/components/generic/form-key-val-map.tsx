import {
  Button,
  ButtonGroup,
  Col,
  FormGroup,
  Row,
  Stack,
} from "react-bootstrap";
import { FC, memo, useCallback, useEffect, useMemo } from "react";
import { faMinus, faPlus } from "@fortawesome/free-solid-svg-icons";
import { useFieldArray, useFormContext, useWatch } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import FormKeyVal from "./form-key-val";
import { FormLabel } from "components/generic/form";
import { HeaderType } from "types/config";
import { diffObjects } from "utils/diff-objects";
import { isEmptyArray } from "utils";

interface Props {
  name: string;
  label?: string;
  tooltip?: string;
  keyPlaceholder?: string;
  valuePlaceholder?: string;

  defaults?: HeaderType[];
}

/**
 * Returns the form fields for a key-value map
 *
 * @param name - The name of the field in the form
 * @param label - The label for the field
 * @param tooltip - The tooltip for the field
 * @param keyPlaceholder - The placeholder for the key field
 * @param valuePlaceholder - The placeholder for the value field
 * @param defaults - The default values for the field
 * @returns The form fields for a key-value map at name with a label and tooltip
 */
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

  // keep track of the array values so we can switch defaults when they're unchanged
  const fieldValues: HeaderType[] = useWatch({ name: name });
  // useDefaults when the fieldValues are undefined or the same as the defaults
  const useDefaults = useMemo(
    () =>
      isEmptyArray(defaults) ? false : !diffObjects(fieldValues, defaults),
    [fieldValues, defaults]
  );
  // trigger validation on change of defaults being used/not
  useEffect(() => {
    trigger(name);

    // Give the defaults back if the field is empty
    if ((fieldValues ?? []).length === 0)
      defaults?.forEach(() => {
        addItem();
      });
  }, [useDefaults]);

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
              disabled={fields.length === 0}
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
              removeMe={
                // Give the remove that's disabled if there's only one item and it matches the defaults
                fieldValues?.length === 1 ? removeLast : () => remove(index)
              }
              keyPlaceholder={keyPlaceholder}
              valuePlaceholder={valuePlaceholder}
            />
          </Row>
        ))}
      </Stack>
    </FormGroup>
  );
};

export default memo(FormKeyValMap);
