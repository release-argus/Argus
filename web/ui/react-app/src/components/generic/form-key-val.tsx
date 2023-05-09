import { Button, Col } from "react-bootstrap";
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
    <FormItem
      name={`${name}.key`}
      unique
      required
      col_sm={6}
      defaultVal={defaults?.key}
      placeholder={keyPlaceholder}
      onMiddle
    />
    <FormItem
      name={`${name}.value`}
      required
      col_xs={10}
      col_sm={5}
      defaultVal={defaults?.value}
      placeholder={valuePlaceholder}
      onRight
    />
  </>
);

export default memo(FormKeyVal);
