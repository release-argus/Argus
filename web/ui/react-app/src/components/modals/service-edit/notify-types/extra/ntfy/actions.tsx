import {
  Button,
  ButtonGroup,
  Col,
  FormGroup,
  Row,
  Stack,
} from "react-bootstrap";
import { FC, memo, useCallback, useEffect, useMemo, useState } from "react";
import { faMinus, faPlus } from "@fortawesome/free-solid-svg-icons";
import { useFieldArray, useFormContext, useWatch } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormLabel } from "components/generic/form";
import NtfyAction from "./action";
import { convertNtfyActionsFromString } from "components/modals/service-edit/util/api-ui-conversions";
import { diffObjects } from "utils/diff-objects";

interface Props {
  name: string;
  label: string;
  tooltip: string;
  defaults?: string;
}

const NtfyActions: FC<Props> = ({ name, label, tooltip, defaults }) => {
  const { setValue, trigger } = useFormContext();
  const { fields, append, remove } = useFieldArray({
    name: name,
  });
  const addItem = useCallback(() => {
    append(
      {
        action: "view",
        label: "",
        url: "",
        intent: "io.heckel.ntfy.USER_ACTION",
      },
      { shouldFocus: false }
    );
  }, []);
  const removeLast = useCallback(() => {
    remove(fields.length - 1);
  }, [fields.length]);

  const actionDefaults = useMemo(
    () => (defaults ? convertNtfyActionsFromString(defaults) : undefined),
    [defaults]
  );
  // keep track of the array values so we can switch defaults when they're unchanged
  const fieldValues = useWatch({ name: name });
  // useDefaults when the fieldValues are undefined or the same as the defaults
  const [useDefaults, setUseDefaults] = useState(false);
  // useEffect rather than useMemo as it was sometimes giving partial fieldValues on load
  useEffect(() => {
    const result = diffObjects(fieldValues, actionDefaults)
    if (result != useDefaults) setUseDefaults(result);
  }, [fieldValues, actionDefaults]);
  // trigger validation on change of defaults being used/not
  useEffect(() => {
    trigger(name);
  }, [useDefaults]);

  // on load, ensure we don't have another types actions
  // and give the defaults if not overridden
  useEffect(() => {
    let values = fieldValues ?? [];
    // ensure we don't have another types actions
    for (const item of values) {
      if ((item.action || "") === "") {
        setValue(name, []);
        values = [];
        break;
      }
    }

    if (values.length === 0) {
      actionDefaults?.forEach((dflt) => {
        append({ action: dflt.action }, { shouldFocus: false });
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
      <Stack>
        {fields.map(({ id }, index) => (
          <Row key={id}>
            <NtfyAction
              name={`${name}.${index}`}
              removeMe={() => remove(index)}
              defaults={useDefaults ? actionDefaults?.[index] : undefined}
            />
          </Row>
        ))}
      </Stack>
    </FormGroup>
  );
};

export default memo(NtfyActions);
