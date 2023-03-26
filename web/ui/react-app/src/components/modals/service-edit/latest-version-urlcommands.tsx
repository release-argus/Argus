/* eslint-disable react/prop-types */
import { Button, ButtonGroup, Col, Row } from "react-bootstrap";
import { faMinus, faPlus } from "@fortawesome/free-solid-svg-icons";
import { useFieldArray, useFormContext } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormLabel } from "components/generic/form";
import FormURLCommand from "./latest-version-urlcommand";
import { memo } from "react";

const FormURLCommands = () => {
  const { control } = useFormContext();
  const { fields, append, remove } = useFieldArray({
    control,
    name: "latest_version.url_commands",
  });

  return (
    <>
      <span className={"pt-1"}>
        <FormLabel text="URL Commands" />
      </span>
      {fields.map(({ id }, index) => {
        return (
          <Row key={id} className="d-flex align-items-center">
            <FormURLCommand
              name={`latest_version.url_commands.${index}`}
              removeMe={() => remove(index)}
            />
          </Row>
        );
      })}
      <Row>
        <Col>
          <ButtonGroup style={{ float: "right" }}>
            <Button
              aria-label="Add new URL Command"
              className="btn-unchecked mb-3"
              variant="success"
              style={{ float: "right" }}
              onClick={() =>
                append(
                  {
                    type: "regex",
                    regex: "",
                    text: "",
                    index: 0,
                    old: "",
                    new: "",
                  },
                  { shouldFocus: false }
                )
              }
            >
              <FontAwesomeIcon icon={faPlus} />
            </Button>
            <Button
              aria-label="Remove last URL Command"
              className="btn-unchecked mb-3"
              variant="danger"
              style={{ float: "left" }}
              onClick={() => remove(fields.length - 1)}
            >
              <FontAwesomeIcon icon={faMinus} />
            </Button>
          </ButtonGroup>
        </Col>
      </Row>
    </>
  );
};

export default memo(FormURLCommands);
