import {
  Button,
  ButtonGroup,
  Col,
  FormGroup,
  Row,
  Stack,
} from "react-bootstrap";
import { FC, memo, useCallback, useEffect, useMemo } from "react";
import { HeaderType, NotifyNtfyAction } from "types/config";
import { faMinus, faPlus } from "@fortawesome/free-solid-svg-icons";
import { useFieldArray, useFormContext, useWatch } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormLabel } from "components/generic/form";
import NtfyAction from "./action";
import { diffObjects } from "utils/diff-objects";

interface Props {
  name: string;
  label: string;
  tooltip: string;
  defaults?: NotifyNtfyAction[];
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

  // keep track of the array values so we can switch defaults when they're unchanged
  const fieldValues = useWatch({ name: name });
  // useDefaults when the fieldValues are unset or the same as the defaults
  const useDefaults = useMemo(
    () => fieldValues && defaults && diffObjects(fieldValues, defaults),
    [fieldValues, defaults]
  );
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
      defaults?.forEach((dflt) => {
        if (dflt.action === undefined) return;
        // http
        else if (dflt.action === "http") {
          const headers = ((dflt?.headers ?? []) as HeaderType[]).map(
            () => ({})
          );
          append(
            { action: dflt.action, headers: headers },
            { shouldFocus: false }
          );
          // broadcast
        } else if (dflt.action === "broadcast") {
          const extras = ((dflt?.extras ?? []) as HeaderType[]).map(() => ({}));
          append(
            { action: dflt.action, extras: extras },
            { shouldFocus: false }
          );
          // view
        } else append({ action: dflt.action }, { shouldFocus: false });
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
              defaults={useDefaults ? defaults?.[index] : undefined}
            />
          </Row>
        ))}
      </Stack>
    </FormGroup>
  );
};

export default memo(NtfyActions);
