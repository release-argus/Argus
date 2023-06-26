import { Button, ButtonGroup, Col, FormGroup, Row } from "react-bootstrap";
import { FC, memo, useCallback, useEffect, useMemo } from "react";
import { faMinus, faPlus } from "@fortawesome/free-solid-svg-icons";
import { useFieldArray, useFormContext, useWatch } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormLabel } from "components/generic/form";
import { NotifyOpsGenieTarget } from "types/config";
import OpsGenieTarget from "./target";
import { diffObjects } from "utils/diff-objects";

interface Props {
  name: string;
  label: string;
  tooltip: string;

  defaults?: NotifyOpsGenieTarget[];
}

const OpsGenieTargets: FC<Props> = ({ name, label, tooltip, defaults }) => {
  const { trigger } = useFormContext();
  const { fields, append, remove } = useFieldArray({
    name: name,
  });
  const addItem = useCallback(() => {
    append({ type: "team", sub_type: "id", value: "" }, { shouldFocus: false });
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
      defaults?.forEach((dflt) => {
        append(
          { type: dflt.type, sub_type: dflt.sub_type },
          { shouldFocus: false }
        );
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
            >
              <FontAwesomeIcon icon={faMinus} />
            </Button>
          </ButtonGroup>
        </Col>
      </Row>
      {fields.map(({ id }, index) => (
        <Row key={id}>
          <OpsGenieTarget
            name={`${name}.${index}`}
            removeMe={() => remove(index)}
            defaults={useDefaults ? defaults?.[index] : undefined}
          />
        </Row>
      ))}
    </FormGroup>
  );
};

export default memo(OpsGenieTargets);
