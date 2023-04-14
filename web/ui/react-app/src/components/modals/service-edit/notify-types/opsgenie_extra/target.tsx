import { Button, Col } from "react-bootstrap";
import { FC, memo } from "react";
import { FormItem, FormSelect } from "components/generic/form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faTrash } from "@fortawesome/free-solid-svg-icons";
import { useWatch } from "react-hook-form";

interface Props {
  name: string;
  removeMe: () => void;
}

const OpsGenieTarget: FC<Props> = ({ name, removeMe }) => {
  const targetTypes = [
    { label: "Team", value: "team" },
    { label: "User", value: "user" },
  ];

  const targetType = useWatch({ name: `${name}.type` });

  return (
    <>
      <FormSelect
        name={`${name}.type`}
        col_xs={6}
        col_sm={3}
        options={targetTypes}
      />
      <FormSelect
        name={`${name}.sub_type`}
        col_xs={6}
        col_sm={3}
        options={[
          { label: "ID", value: "id" },
          targetType === "team"
            ? { label: "Name", value: "name" }
            : { label: "Username", value: "username" },
        ]}
        onMiddle
      />
      <FormItem
        name={`${name}.value`}
        required
        col_xs={11}
        col_sm={5}
        placeholder="<value>"
        onRight
      />
      <Col xs={1}>
        <Button
          className="btn-unchecked btn-icon-center"
          style={{ paddingTop: "0.94rem" }}
          onClick={removeMe}
        >
          <FontAwesomeIcon icon={faTrash} />
        </Button>
      </Col>
    </>
  );
};

export default memo(OpsGenieTarget);
