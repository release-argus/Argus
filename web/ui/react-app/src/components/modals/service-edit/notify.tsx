import { Accordion, Button, Col, Form, FormGroup, Row } from "react-bootstrap";
import {
  Dict,
  LatestVersionLookupType,
  NotifyType,
  NotifyTypes,
  NotifyTypesConst,
} from "types/config";
import { FC, JSX, memo, useEffect, useMemo } from "react";
import { FormItem, FormLabel, FormSelect } from "components/generic/form";
import { useFormContext, useWatch } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { NotifyEditType } from "types/service-edit";
import RenderNotify from "./notify-types/render";
import { TYPE_OPTIONS } from "./notify-types/types";
import TestNotify from "components/modals/service-edit/test-notify";
import { convertNotifyToAPI } from "components/modals/service-edit/util";
import { faTrash } from "@fortawesome/free-solid-svg-icons";

interface Props {
  name: string;
  removeMe: () => void;

  serviceName: string;
  originals?: NotifyEditType[];
  globalNotifyOptions: JSX.Element;
  globals?: Dict<NotifyType>;
  defaults?: Dict<NotifyType>;
  hard_defaults?: Dict<NotifyType>;
}

/**
 * Notify is the form fields for a Notify
 *
 * @param name - The name of the field in the form
 * @param removeMe - The function to remove this Notify
 * @param serviceName - The name of the service
 * @param originals - The original values for the Notify
 * @param globalNotifyOptions - The options for the global Notify's
 * @param globals - The global Notify's
 * @param defaults - The default values for all Notify types
 * @param hard_defaults - The hard default values for all Notify types
 * @returns The form fields for this Notify
 */
const Notify: FC<Props> = ({
  name,
  removeMe,

  serviceName,
  originals,
  globalNotifyOptions,
  globals,
  defaults,
  hard_defaults,
}) => {
  const { setValue, trigger } = useFormContext();

  const itemName: string = useWatch({ name: `${name}.name` });
  const itemType: NotifyTypes = useWatch({ name: `${name}.type` });
  const lvType: LatestVersionLookupType["type"] = useWatch({
    name: "latest_version.type",
  });
  const lvURL: string | undefined = useWatch({ name: "latest_version.url" });
  const webURL: string | undefined = useWatch({ name: "dashboard.web_url" });
  useEffect(() => {
    // Set Type to that of the global for the new name if it exists
    if (globals?.[itemName]?.type)
      setValue(`${name}.type`, globals[itemName].type);
    else if ((itemType ?? "") === "" && NotifyTypesConst.includes(itemName))
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
  const original = useMemo(() => {
    if (originals) {
      for (const o of originals) {
        if (o.oldIndex === itemName) {
          return convertNotifyToAPI(o);
        }
      }
    }
    return { options: {}, url_fields: {}, params: {} };
  }, [originals]);
  const serviceURL =
    lvType === "github" && (lvURL?.match(/\//g) ?? []).length > 1
      ? `https://github.com/${lvURL}`
      : lvURL;

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
                  globals && Object.keys(globals).indexOf(itemName) !== -1
                    ? itemName
                    : ""
                }
                onChange={(e) => setValue(`${name}.name`, e.target.value)}
              >
                {globalNotifyOptions}
              </Form.Select>
            </FormGroup>
          </Col>
          <FormSelect
            name={`${name}.type`}
            customValidation={(value) => {
              if (
                itemType !== undefined &&
                globals?.[itemName]?.type &&
                itemType !== globals?.[itemName]?.type
              ) {
                return `${value} does not match the global for "${itemName}" of ${globals?.[itemName]?.type}. Either change the type to match that, or choose a new name`;
              }
              return true;
            }}
            col_xs={6}
            label="Type"
            options={TYPE_OPTIONS}
            position="right"
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
            globalNotify={globals?.[itemName]}
            defaults={defaults?.[itemType]}
            hard_defaults={hard_defaults?.[itemType]}
          />
          <TestNotify
            path={name}
            original={original}
            extras={{
              service_name: serviceName,
              service_url: serviceURL,
              web_url: webURL,
            }}
          />
        </Row>
      </Accordion.Body>
    </Accordion>
  );
};

export default memo(Notify);
