import { Button, Col, Row } from "react-bootstrap";
import { FC, memo } from "react";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormItem } from "components/generic/form";
import { HeaderType } from "types/config";
import { faTrash } from "@fortawesome/free-solid-svg-icons";

interface Props {
  name: string;
  defaults?: HeaderType;
  removeMe: () => void;
  keyPlaceholder?: string;
  valuePlaceholder?: string;
}

/**
 * Returns the form fields for a key-value pair
 *
 * @param name - The name of the field in the form
 * @param defaults - The default values for the field
 * @param removeMe - The function to remove the field
 * @param keyPlaceholder - The placeholder for the key field
 * @param valuePlaceholder - The placeholder for the value field
 * @returns The form fields for a key-value pair at name
 */
const FormKeyVal: FC<Props> = ({
  name,
  defaults,
  removeMe,
  keyPlaceholder = "e.g. X-Header",
  valuePlaceholder = "e.g. value",
}) => (
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
    <Col>
      <Row>
        <FormItem
          name={`${name}.key`}
          required
          unique
          col_sm={6}
          defaultVal={defaults?.key}
          placeholder={keyPlaceholder}
          position="middle"
        />
        <FormItem
          name={`${name}.value`}
          required
          col_sm={6}
          defaultVal={defaults?.value}
          placeholder={valuePlaceholder}
          position="right"
        />
      </Row>
    </Col>
  </>
);

export default memo(FormKeyVal);
