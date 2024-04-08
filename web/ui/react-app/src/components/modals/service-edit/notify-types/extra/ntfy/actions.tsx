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
import { isEmptyArray, isEmptyOrNull } from "utils";
import { useFieldArray, useFormContext, useWatch } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormLabel } from "components/generic/form";
import { NotifyNtfyAction } from "types/config";
import NtfyAction from "./action";
import { convertNtfyActionsFromString } from "components/modals/service-edit/util";
import { diffObjects } from "utils/diff-objects";

interface Props {
  name: string;
  label: string;
  tooltip: string;
  defaults?: NotifyNtfyAction[];
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

  // keep track of the array values so we can switch defaults when they're unchanged
  const fieldValues: NotifyNtfyAction[] = useWatch({ name: name });
  // useDefaults when the fieldValues are unset or the same as the defaults
  const useDefaults = useMemo(
    () =>
      isEmptyArray(defaults)
        ? false
        : !diffObjects(fieldValues, defaults, [".action"]),
    [fieldValues, defaults]
  );

  // Keep only selects/length of arrays
  const trimmedDefaults = useMemo(
    () => convertNtfyActionsFromString(undefined, JSON.stringify(defaults)),
    [defaults]
  );
  // trigger validation on change of defaults being used/not
  useEffect(() => {
    trigger(name);

    // Give the defaults back if the field is empty
    if (isEmptyArray(fieldValues)) {
      trimmedDefaults.forEach((dflt) => {
        append(dflt, { shouldFocus: false });
      });
    }
  }, [useDefaults]);

  // on load, ensure we don't have another types actions
  // and give the defaults if not overridden
  useEffect(() => {
    // ensure we don't have another types actions
    for (const item of fieldValues) {
      if (isEmptyOrNull(item.action)) {
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
            <NtfyAction
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

export default memo(NtfyActions);
