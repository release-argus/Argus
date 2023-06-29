import { Accordion, Button, Col, Form, FormGroup, Row } from "react-bootstrap";
import { Dict, WebHookType } from "types/config";
import { FC, useEffect, useMemo } from "react";
import {
  FormItem,
  FormKeyValMap,
  FormLabel,
  FormSelect,
} from "components/generic/form";
import { useFormContext, useWatch } from "react-hook-form";

import { BooleanWithDefault } from "components/generic";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faTrash } from "@fortawesome/free-solid-svg-icons";
import { globalOrDefault } from "./notify-types/util";

interface Props {
  name: string;
  removeMe: () => void;

  globalOptions: JSX.Element;
  globals?: Dict<WebHookType>;
  defaults?: WebHookType;
  hard_defaults?: WebHookType;
}

const EditServiceWebHook: FC<Props> = ({
  name,
  removeMe,

  globalOptions,
  globals,
  defaults,
  hard_defaults,
}) => {
  const webHookTypeOptions = [
    { label: "GitHub", value: "github" },
    { label: "GitLab", value: "gitlab" },
  ];

  const { setValue, trigger } = useFormContext();

  const itemName = useWatch({ name: `${name}.name` });
  const itemType = useWatch({ name: `${name}.type` });
  const global = globals && globals[itemName];
  useEffect(() => {
    global?.type && setValue(`${name}.type`, global.type);
  }, [global]);
  useEffect(() => {
    if (globals?.[itemName]?.type !== undefined)
      setValue(`${name}.type`, globals[itemName].type);
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
              <FormLabel text="Global?" tooltip="Use this WebHook as a base" />
              <Form.Select
                value={
                  globals
                    ? itemName !== "" &&
                      Object.keys(globals).indexOf(itemName) !== -1
                      ? itemName
                      : ""
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
                globals?.[itemName]?.type &&
                itemType !== globals?.[itemName]?.type
              ) {
                return `${value} does not match the global for "${itemName}" of ${globals?.[itemName]?.type}. Either change the type to match that, or choose a new name`;
              }
              return true;
            }}
            col_xs={6}
            label="Type"
            tooltip="Style of WebHook to emulate"
            options={webHookTypeOptions}
            onRight
          />
          <FormItem
            name={`${name}.name`}
            required
            unique
            col_sm={12}
            label={"Name"}
            onRight
          />
          <FormItem
            name={`${name}.url`}
            required
            col_sm={12}
            type="text"
            label="Target URL"
            tooltip="Where to send the WebHook"
            defaultVal={globalOrDefault(
              global?.url,
              defaults?.url,
              hard_defaults?.url
            )}
            isURL
          />
          <BooleanWithDefault
            name={`${name}.allow_invalid_certs`}
            label="Allow Invalid Certs"
            defaultValue={
              global?.allow_invalid_certs ||
              defaults?.allow_invalid_certs ||
              hard_defaults?.allow_invalid_certs
            }
          />
          <FormItem
            name={`${name}.secret`}
            required
            col_sm={12}
            label="Secret"
            defaultVal={
              global?.secret || defaults?.secret || hard_defaults?.secret
            }
          />
          <FormKeyValMap name={`${name}.custom_headers`} />
          <FormItem
            name={`${name}.desired_status_code`}
            col_xs={6}
            label="Desired Status Code"
            tooltip="Treat the WebHook as successful when this status code is received (0=2XX)"
            defaultVal={globalOrDefault(
              global?.desired_status_code,
              defaults?.desired_status_code,
              hard_defaults?.desired_status_code
            )}
          />
          <FormItem
            name={`${name}.max_tries`}
            col_xs={6}
            label="Max tries"
            defaultVal={`${
              global?.max_tries ||
              defaults?.max_tries ||
              hard_defaults?.max_tries ||
              ""
            }`}
            onRight
          />
          <FormItem
            name={`${name}.delay`}
            col_sm={12}
            label="Delay"
            tooltip="Delay sending by this duration"
            defaultVal={
              global?.delay || defaults?.delay || hard_defaults?.delay
            }
            onRight
          />
          <BooleanWithDefault
            name={`${name}.silent_fails`}
            label="Silent fails"
            tooltip="Notify if WebHook fails max tries times"
            defaultValue={
              global?.silent_fails ||
              defaults?.silent_fails ||
              hard_defaults?.silent_fails
            }
          />
        </Row>
      </Accordion.Body>
    </Accordion>
  );
};

export default EditServiceWebHook;
