/* eslint-disable react/prop-types */
import { Button, ButtonGroup, Col, Row } from "react-bootstrap";
import { faMinus, faPlus } from "@fortawesome/free-solid-svg-icons";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormLabel } from "components/generic/form";
import FormURLCommand from "./latest-version-urlcommand";
import { memo } from "react";
import { useFieldArray } from "react-hook-form";

const FormURLCommands = () => {
  const { fields, append, remove } = useFieldArray({
    name: "latest_version.url_commands",
  });

  return (
    <>
      <Row>
        <Col>
          <FormLabel text="URL Commands" />
        </Col>
        <Col>
          <ButtonGroup style={{ float: "right" }}>
            <Button
              aria-label="Add new URL Command"
              className="btn-unchecked"
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
              className="btn-unchecked"
              variant="danger"
              style={{ float: "left" }}
              onClick={() => remove(fields.length - 1)}
            >
              <FontAwesomeIcon icon={faMinus} />
            </Button>
          </ButtonGroup>
        </Col>
      </Row>
      {fields.map(({ id }, i, { length }) => {
        return (
          <Row
            key={id}
            className={`d-flex align-items-center ${
              length - 1 === i ? "mb-2" : ""
            }`}
          >
            <FormURLCommand
              name={`latest_version.url_commands.${i}`}
              removeMe={() => remove(i)}
            />
          </Row>
        );
      })}
      {fields.length !== 0 && <br />}
    </>
  );
};

export default memo(FormURLCommands);
