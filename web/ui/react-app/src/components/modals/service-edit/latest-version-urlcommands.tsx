import { Button, ButtonGroup, Col, Row } from "react-bootstrap";
import { faMinus, faPlus } from "@fortawesome/free-solid-svg-icons";
import { memo, useCallback } from "react";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormLabel } from "components/generic/form";
import FormURLCommand from "./latest-version-urlcommand";
import { isEmptyArray } from "utils";
import { useFieldArray } from "react-hook-form";

/**
 * @returns The form fields for a list of `latest_version.url_commands`
 */
const FormURLCommands = () => {
  const { fields, append, remove } = useFieldArray({
    name: "latest_version.url_commands",
  });

  const addItem = useCallback(() => {
    append(
      {
        type: "regex",
        regex: "",
        text: "",
        index: null,
        old: "",
        new: "",
      },
      { shouldFocus: false },
    );
  }, []);
  const removeLast = useCallback(() => {
    remove(fields.length - 1);
  }, [fields.length]);

  return (
    <>
      <Row>
        <Col className="pt-1">
          <FormLabel text="URL Commands" />
        </Col>
        <Col>
          <ButtonGroup style={{ float: "right" }}>
            <Button
              aria-label="Add new URL Command"
              className="btn-unchecked"
              variant="success"
              style={{ float: "right" }}
              onClick={addItem}
            >
              <FontAwesomeIcon icon={faPlus} />
            </Button>
            <Button
              aria-label="Remove last URL Command"
              className="btn-unchecked"
              variant="danger"
              style={{ float: "left" }}
              onClick={removeLast}
              disabled={isEmptyArray(fields)}
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
      {!isEmptyArray(fields) && <br />}
    </>
  );
};

export default memo(FormURLCommands);
