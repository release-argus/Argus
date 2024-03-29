import { Button, Col } from "react-bootstrap";
import { FC, memo } from "react";
import { FormItem, FormSelect } from "components/generic/form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { NotifyOpsGenieTarget } from "types/config";
import { faTrash } from "@fortawesome/free-solid-svg-icons";
import { useWatch } from "react-hook-form";

interface Props {
  name: string;
  removeMe: () => void;

  defaults?: NotifyOpsGenieTarget;
}

const OpsGenieTarget: FC<Props> = ({ name, removeMe, defaults }) => {
  const targetTypes = [
    { label: "Team", value: "team" },
    { label: "User", value: "user" },
  ];

  const targetType = useWatch({ name: `${name}.type` });

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
        defaultVal={defaults?.value}
        onRight
      />
    </>
  );
};

export default memo(OpsGenieTarget);
