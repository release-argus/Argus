import { Button, Col } from "react-bootstrap";
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
      <FormSelect
        col_xs={3}
        col_sm={3}
        name={`${name}.type`}
        label="Type"
        smallLabel
        options={urlCommandTypeOptions}
      />
      {commandType && (
        <RenderURLCommand name={name} commandType={commandType} />
      )}
      <Col xs={1} className="d-flex align-items-center justify-content-end">
        <Button
          aria-label="Remove URL Command"
          className="btn-unchecked"
          variant="warning"
          onClick={removeMe}
        >
          <FontAwesomeIcon icon={faTrash} />
        </Button>
      </Col>
    </>
  );
};

export default memo(FormURLCommand);
