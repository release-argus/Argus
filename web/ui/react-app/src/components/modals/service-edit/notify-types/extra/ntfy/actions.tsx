import { Button, ButtonGroup, Col, Form, Row, Stack } from "react-bootstrap";
import { FC, memo, useCallback, useEffect, useMemo } from "react";
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
  const { trigger } = useFormContext();
  const { fields, append, remove } = useFieldArray({
    name: name,
  });
  const addItem = useCallback(() => {
    append({ action: "view" }, { shouldFocus: false });
  }, []);
  const removeLast = useCallback(() => {
    remove(fields.length - 1);
  }, [fields]);

  const actionDefaults = useMemo(
    () => (defaults ? convertNtfyActionsFromString(defaults) : undefined),
    [defaults]
  );
  // keep track of the array values so we can switch defaults when they're unchanged
  const fieldValues = useWatch({ name: name });
  // useDefaults when the fieldValues are undefined or the same as the defaults
  const useDefaults = useMemo(
    () => diffObjects(fieldValues, actionDefaults),
    [fieldValues, defaults]
  );
  useEffect(() => {
    trigger(name);
  }, [useDefaults]);

  // on load, give the defaults if not overridden
  useEffect(() => {
    if (useDefaults) {
      actionDefaults?.forEach((dflt) => {
        append(
          { action: dflt.action, label: dflt.label, method: dflt.method },
          { shouldFocus: false }
        );
      });
    }
  }, []);

  return (
    <Form.Group>
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
      <Stack gap={1}>
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
    </Form.Group>
  );
};

export default memo(NtfyActions);
