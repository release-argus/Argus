import { Accordion, Button, Col, Form, Row } from "react-bootstrap";
import { FC, memo, useEffect, useMemo } from "react";
import { FormItem, FormLabel, FormSelect } from "components/generic/form";
import { NotifyType, ServiceDict } from "types/config";
import { useFormContext, useWatch } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import RenderNotify from "./notify-types/render";
import { TYPE_OPTIONS } from "./notify-types/types";
import { faTrash } from "@fortawesome/free-solid-svg-icons";

interface Props {
  name: string;
  removeMe: () => void;

  globalNotifyOptions: JSX.Element;
  globals?: ServiceDict<NotifyType>;
  defaults?: ServiceDict<NotifyType>;
  hard_defaults?: ServiceDict<NotifyType>;
}

const Notify: FC<Props> = ({
  name,
  removeMe,

  globalNotifyOptions,
  globals,
  defaults,
  hard_defaults,
}) => {
  const { setValue, trigger } = useFormContext();

  const itemName = useWatch({ name: `${name}.name` });
  const itemType = useWatch({ name: `${name}.type` });
  useEffect(() => {
    if (globals && itemName !== "" && globals[itemName]?.type !== undefined) {
      setValue(`${name}.type`, globals[itemName].type);
      trigger();
    }
  }, [itemName]);
  const header = useMemo(
    () => `${Number(name.split(".").slice(-1)) + 1}: (${itemType}) ${itemName}`,
    [name, itemName, itemType]
  );

  return (
    <Accordion>
      <div style={{ display: "flex", alignItems: "center" }}>
        <Button
          className="btn-unchecked"
          variant="secondary"
          onClick={removeMe}
        >
          <FontAwesomeIcon icon={faTrash} />
        </Button>
        <Accordion.Button className="p-2">{header}</Accordion.Button>
      </div>

      <Accordion.Body>
        <Row xs={12}>
          <Col xs={6} className={`pe-2 pt-1 pb-1`}>
            <Form.Group className="mb-2">
              <FormLabel text="Global?" tooltip="Use this Notify as a base" />
              <Form.Select
                value={
                  globals &&
                  itemName !== "" &&
                  Object.keys(globals).indexOf(itemName) !== -1
                    ? itemName
                    : ""
                }
                onChange={(e) => setValue(`${name}.name`, e.target.value)}
              >
                {globalNotifyOptions}
              </Form.Select>
            </Form.Group>
          </Col>
          <FormSelect
            name={`${name}.type`}
            col_xs={6}
            label="Type"
            options={TYPE_OPTIONS}
            onRight
          />
          <FormItem
            name={`${name}.name`}
            required
            unique
            col_sm={12}
            label="Name"
          />
          <RenderNotify
            name={name}
            type={itemType}
            globalNotify={globals && itemName ? globals[itemName] : undefined}
            defaults={defaults}
            hard_defaults={hard_defaults && hard_defaults[itemType]}
          />
        </Row>
      </Accordion.Body>
    </Accordion>
  );
};

export default memo(Notify);
