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
import { NotifyNtfyAction } from "types/config";
import NtfyAction from "./action";
import { convertNtfyActionsFromString } from "components/modals/service-edit/util/api-ui-conversions";
import { diffObjects } from "utils/diff-objects";

interface Props {
  name: string;
  label: string;
  tooltip: string;
  defaults?: string;
}

/**
 * NtfyActions is the form fields for a list of Ntfy actions
 *
 * @param name - The name of the field in the form
 * @param label - The label for the field
 * @param tooltip - The tooltip for the field
 * @param defaults - The default values for the field
 * @returns A set of form fields for a list of Ntfy actions
 */
const NtfyActions: FC<Props> = ({ name, label, tooltip, defaults }) => {
  const { trigger } = useFormContext();
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
  }, [fields]);

  const actionDefaults = useMemo(
    () => (defaults ? convertNtfyActionsFromString(defaults) : undefined),
    [defaults]
  );
  // keep track of the array values so we can switch defaults when they're unchanged
  const fieldValues: NotifyNtfyAction[] = useWatch({ name: name });
  // useDefaults when the fieldValues are unset or the same as the defaults
  const useDefaults = useMemo(
    () => diffObjects(fieldValues, actionDefaults),
    [fieldValues, defaults]
  );
  useEffect(() => {
    trigger(name);
  }, [useDefaults]);

  // on load, give the defaults if not overridden
  useEffect(() => {
    if (useDefaults)
      actionDefaults?.forEach((dflt) => {
        append(
          { action: dflt.action, label: dflt.label, method: dflt.method },
          { shouldFocus: false }
        );
      });
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
