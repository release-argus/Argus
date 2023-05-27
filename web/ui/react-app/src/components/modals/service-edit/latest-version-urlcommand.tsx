import { Button, Col, Row } from "react-bootstrap";
import { FC, memo } from "react";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormSelect } from "components/generic/form";
import RenderURLCommand from "./url-commands/render";
import { faTrash } from "@fortawesome/free-solid-svg-icons";
import { useWatch } from "react-hook-form";

interface Props {
  name: string;
  removeMe: () => void;
}

const FormURLCommand: FC<Props> = ({ name, removeMe }) => {
  const urlCommandTypeOptions = [
    { label: "RegEx", value: "regex" },
    { label: "Replace", value: "replace" },
    { label: "Split", value: "split" },
  ];

  const commandType = useWatch({ name: `${name}.type` });

  return (
    <>
      <Col xs={2} sm={1} style={{ height: "100%", padding: "0.25rem" }}>
        <Button
          className="btn-secondary-outlined btn-icon-center"
          variant="secondary"
          onClick={removeMe}
        >
          <FontAwesomeIcon icon={faTrash} />
        </Button>
      </Col>
      <Col>
        <Row>
          <FormSelect
            col_xs={4}
            col_sm={4}
            name={`${name}.type`}
            label="Type"
            smallLabel
            options={urlCommandTypeOptions}
          />
          {commandType && (
            <RenderURLCommand name={name} commandType={commandType} />
          )}
        </Row>
      </Col>
    </>
  );
};

export default memo(FormURLCommand);
