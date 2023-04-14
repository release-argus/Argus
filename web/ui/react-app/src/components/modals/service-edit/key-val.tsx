import { Button, Col } from "react-bootstrap";
import { FC, memo } from "react";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormItem } from "components/generic/form";
import { faTrash } from "@fortawesome/free-solid-svg-icons";

interface Props {
  name: string;
  removeMe: () => void;
  keyPlaceholder?: string;
  valuePlaceholder?: string;
}

const FormKeyVal: FC<Props> = ({
  name,
  removeMe,
  keyPlaceholder = "e.g. X-Header",
  valuePlaceholder = "e.g. value",
}) => (
  <>
    <FormItem
      name={`${name}.key`}
      unique
      required
      col_sm={6}
      placeholder={keyPlaceholder}
    />
    <FormItem
      name={`${name}.value`}
      required
      col_xs={10}
      col_sm={5}
      placeholder={valuePlaceholder}
    />
    <Col
      xs={2}
      sm={1}
      className="d-flex align-items-center justify-content-end"
    >
      <Button className="btn-unchecked" variant="warning" onClick={removeMe}>
        <FontAwesomeIcon icon={faTrash} />
      </Button>
    </Col>
  </>
);

export default memo(FormKeyVal);
