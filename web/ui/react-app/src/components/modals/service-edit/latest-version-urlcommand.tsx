import { Button, Col, Row } from "react-bootstrap";
import { FC, memo } from "react";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormSelect } from "components/generic/form";
import RenderURLCommand from "./url-commands/render";
import { URLCommandTypes } from "types/config";
import { faTrash } from "@fortawesome/free-solid-svg-icons";
import { useWatch } from "react-hook-form";

interface Props {
  name: string;
  removeMe: () => void;
}

/**
 * Returns the form fields for a URL command
 *
 * @param name - The name of the field in the form
 * @param removeMe - The function to remove the command
 * @returns The form fields for this URL command
 */
const FormURLCommand: FC<Props> = ({ name, removeMe }) => {
  const urlCommandTypeOptions: { label: string; value: URLCommandTypes }[] = [
    { label: "RegEx", value: "regex" },
    { label: "Replace", value: "replace" },
    { label: "Split", value: "split" },
  ];

  const commandType: URLCommandTypes = useWatch({ name: `${name}.type` });

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
            col_xs={5}
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
