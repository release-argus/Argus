import { Accordion, Button, Col, Form, FormGroup, Row } from "react-bootstrap";
import { Dict, NotifyType, NotifyTypes, NotifyTypesConst } from "types/config";
import { FC, JSX, memo, useEffect, useMemo } from "react";
import { FormItem, FormLabel, FormSelect } from "components/generic/form";
import { useFormContext, useWatch } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import RenderNotify from "./notify-types/render";
import { TYPE_OPTIONS } from "./notify-types/types";
import { faTrash } from "@fortawesome/free-solid-svg-icons";

interface Props {
  name: string;
  removeMe: () => void;

  globalOptions: JSX.Element;
  mains?: Dict<NotifyType>;
  defaults?: Dict<NotifyType>;
  hard_defaults?: Dict<NotifyType>;
}

/**
 * returns the form fields for a notify
 *
 * @param name - The name of the field in the form
 * @param removeMe - The function to remove this Notify
 * @param globalOptions - The options for the global Notify's
 * @param mains - The main Notify's
 * @param defaults - The default values for all Notify types
 * @param hard_defaults - The hard default values for all Notify types
 * @returns The form fields for this Notify
 */
const Notify: FC<Props> = ({
  name,
  removeMe,

  globalOptions,
  mains,
  defaults,
  hard_defaults,
}) => {
  const { setValue, trigger } = useFormContext();

  const itemName: string = useWatch({ name: `${name}.name` });
  const itemType: NotifyTypes = useWatch({ name: `${name}.type` });
  useEffect(() => {
    // Set Type to that of the global for the new name if it exists
    if (mains?.[itemName]?.type) setValue(`${name}.type`, mains[itemName].type);
    else if (itemType && (NotifyTypesConst as string[]).includes(itemName))
      setValue(`${name}.type`, itemName);
    // Trigger validation on name/type
    setTimeout(() => {
      if (itemName !== "") trigger(`${name}.name`);
      trigger(`${name}.type`);
    }, 25);
  }, [itemName]);
  const header = useMemo(
    () => `${name.split(".").slice(-1)}: (${itemType}) ${itemName}`,
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
            <FormGroup className="mb-2">
              <FormLabel text="Global?" tooltip="Use this Notify as a base" />
              <Form.Select
                value={
                  mains && Object.keys(mains).indexOf(itemName) !== -1
                    ? itemName
                    : ""
                }
                onChange={(e) => setValue(`${name}.name`, e.target.value)}
              >
                {globalOptions}
              </Form.Select>
            </FormGroup>
          </Col>
          <FormSelect
            name={`${name}.type`}
            customValidation={(value) => {
              if (
                itemType !== undefined &&
                mains?.[itemName]?.type &&
                itemType !== mains?.[itemName]?.type
              ) {
                return `${value} does not match the global for "${itemName}" of ${mains?.[itemName]?.type}. Either change the type to match that, or choose a new name`;
              }
              return true;
            }}
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
            main={mains?.[itemName]}
            defaults={defaults?.[itemType]}
            hard_defaults={hard_defaults?.[itemType]}
          />
        </Row>
      </Accordion.Body>
    </Accordion>
  );
};

export default memo(Notify);
