import { Button, Col, Row } from "react-bootstrap";
import { FC, memo, useEffect } from "react";
import { FormItem, FormSelect } from "components/generic/form";
import { useFormContext, useWatch } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { NotifyNtfyAction } from "types/config";
import RenderAction from "./render";
import { faTrash } from "@fortawesome/free-solid-svg-icons";

interface Props {
  name: string;
  defaults?: NotifyNtfyAction;
  removeMe: () => void;
}

/**
 * NtfyAction is the form fields for a Ntfy action
 *
 * @param name - The name of the field in the form
 * @param defaults - The default values for the action
 * @param removeMe - The function to remove this action
 * @returns The form fields for this action
 */
const NtfyAction: FC<Props> = ({ name, defaults, removeMe }) => {
  const { setValue } = useFormContext();
  const typeOptions = [
    { label: "View", value: "view" },
    { label: "HTTP", value: "http" },
    { label: "Broadcast", value: "broadcast" },
  ];
  enum typeLabelMap {
    view = "Open page",
    http = "Close door",
    broadcast = "Take picture",
  }

  const targetType: keyof typeof typeLabelMap = useWatch({
    name: `${name}.action`,
  });

  // Set Select's to the defaults
  useEffect(() => {
    if (defaults !== undefined) setValue(`${name}.action`, defaults.action);
    if (defaults?.method !== undefined)
      setValue(`${name}.method`, defaults.method);
  }, []);

  return (
    <>
      <Col xs={2} sm={1} style={{ padding: "0.25rem" }}>
        <Button
          className="btn-secondary-outlined btn-icon-center"
          variant="secondary"
          onClick={removeMe}
        >
          <FontAwesomeIcon icon={faTrash} />
        </Button>
      </Col>
      <Col xs={10} sm={11}>
        <Row>
          <FormSelect
            name={`${name}.action`}
            col_xs={6}
            col_sm={3}
            label="Action Type"
            options={typeOptions}
          />
          <FormItem
            name={`${name}.label`}
            label="Label"
            tooltip="Button name to display on the notification"
            required
            col_xs={6}
            col_sm={4}
            defaultVal={defaults?.label}
            placeholder={`e.g. '${typeLabelMap[targetType]}'`}
            onMiddle
          />
          <RenderAction
            name={name}
            targetType={targetType}
            defaults={defaults}
          />
        </Row>
      </Col>
    </>
  );
};

export default memo(NtfyAction);
