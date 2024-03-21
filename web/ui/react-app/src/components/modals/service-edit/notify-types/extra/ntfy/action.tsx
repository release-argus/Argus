import { Button, Col, Row } from "react-bootstrap";
import { FC, memo } from "react";
import { FormItem, FormSelect } from "components/generic/form";
import { NotifyNtfyAction, NotifyNtfyActionTypes } from "types/config";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import RenderAction from "./render";
import { faTrash } from "@fortawesome/free-solid-svg-icons";
import { useWatch } from "react-hook-form";

interface Props {
  name: string;
  defaults?: NotifyNtfyAction;
  removeMe: () => void;
}

const NtfyAction: FC<Props> = ({ name, defaults, removeMe }) => {
  const typeOptions: { label: string; value: NotifyNtfyActionTypes }[] = [
    { label: "View", value: "view" },
    { label: "HTTP", value: "http" },
    { label: "Broadcast", value: "broadcast" },
  ];
  enum typeLabelMap {
    view = "Open page",
    http = "Close door",
    broadcast = "Take picture",
  }

  const targetType = useWatch({
    name: `${name}.action`,
  }) as keyof typeof typeLabelMap;

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
