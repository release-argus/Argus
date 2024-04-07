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
import { FormLabel } from "components/generic/form";
import { NotifyOpsGenieTarget } from "types/config";
import OpsGenieTarget from "./target";
import { diffObjects } from "utils/diff-objects";
import { isEmptyArray } from "utils";

interface Props {
  name: string;
  label: string;
  tooltip: string;

  defaults?: NotifyOpsGenieTarget[];
}

/**
 * OpsGenieTargets is the form fields for a list of OpsGenie targets
 *
 * @param name - The name of the field in the form
 * @param label - The label for the field
 * @param tooltip - The tooltip for the field
 * @param defaults - The default values for the field
 * @returns A set of form fields for a list of OpsGenie targets
 */
const OpsGenieTargets: FC<Props> = ({ name, label, tooltip, defaults }) => {
  const { trigger } = useFormContext();
  const { fields, append, remove } = useFieldArray({
    name: name,
  });
  const addItem = useCallback(() => {
    append(
      {
        type: "team",
        sub_type: "id",
        value: "",
      },
      { shouldFocus: false }
    );
  }, []);

  // keep track of the array values so we can switch defaults when they're unchanged
  const fieldValues: NotifyOpsGenieTarget[] = useWatch({ name: name });
  // useDefaults when the fieldValues are undefined or the same as the defaults
  const useDefaults = useMemo(
    () =>
      isEmptyArray(defaults)
        ? false
        : !diffObjects(fieldValues, defaults, [".type", ".sub_type"]),
    [fieldValues, defaults]
  );
  useEffect(() => {
    trigger(name);

    // Give the defaults back if the field is empty
    if (fieldValues?.length === 0)
      defaults?.forEach((dflt) => {
        append(
          { type: dflt.type, sub_type: dflt.sub_type, value: "" },
          { shouldFocus: false }
        );
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
              className="btn-unchecked"
              style={{ float: "right" }}
              onClick={addItem}
            >
              <FontAwesomeIcon icon={faPlus} />
            </Button>
            <Button
              aria-label={`Remove last ${label}`}
              className="btn-unchecked"
              style={{ float: "left" }}
              onClick={removeLast}
              disabled={fields.length === 0}
            >
              <FontAwesomeIcon icon={faMinus} />
            </Button>
          </ButtonGroup>
        </Col>
      </Row>
      <Stack>
        {fields.map(({ id }, index) => (
          <Row key={id}>
            <OpsGenieTarget
              name={`${name}.${index}`}
              removeMe={
                // Give the remove that's disabled if there's only one item and it matches the defaults
                fieldValues?.length === 1 ? removeLast : () => remove(index)
              }
              defaults={useDefaults ? defaults?.[index] : undefined}
            />
          </Row>
        ))}
      </Stack>
    </FormGroup>
  );
};

export default memo(OpsGenieTargets);
