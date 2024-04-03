import { Accordion, Button, Col, Form, FormGroup, Row } from "react-bootstrap";
import { Dict, WebHookType } from "types/config";
import { FC, JSX, useEffect, useMemo } from "react";
import {
  FormItem,
  FormKeyValMap,
  FormLabel,
  FormSelect,
} from "components/generic/form";
import { firstNonDefault, firstNonEmpty } from "utils";
import { useFormContext, useWatch } from "react-hook-form";

import { BooleanWithDefault } from "components/generic";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faTrash } from "@fortawesome/free-solid-svg-icons";

interface Props {
  name: string;
  removeMe: () => void;

  globalOptions: JSX.Element;
  mains?: Dict<WebHookType>;
  defaults?: WebHookType;
  hard_defaults?: WebHookType;
}

/**
 * Returns the form fields for a WebHook
 *
 * @param name - The name of the field in the form
 * @param removeMe - The function to remove this WebHook
 * @param globalOptions - The options for the global WebHook's
 * @param mains - The main WebHook's
 * @param defaults - The default values for a WebHook
 * @param hard_defaults - The hard default values for a WebHook
 * @returns The form fields for this WebHook
 */
const EditServiceWebHook: FC<Props> = ({
  name,
  removeMe,

  globalOptions,
  mains,
  defaults,
  hard_defaults,
}) => {
  const webHookTypeOptions: {
    label: string;
    value: NonNullable<WebHookType["type"]>;
  }[] = [
    { label: "GitHub", value: "github" },
    { label: "GitLab", value: "gitlab" },
  ];

  const { setValue, trigger } = useFormContext();

  const itemName: string = useWatch({ name: `${name}.name` });
  const itemType: WebHookType["type"] = useWatch({ name: `${name}.type` });
  const main = mains?.[itemName];
  useEffect(() => {
    main?.type && setValue(`${name}.type`, main.type);
  }, [main]);
  useEffect(() => {
    if (mains?.[itemName]?.type !== undefined)
      setValue(`${name}.type`, mains[itemName].type);
    setTimeout(() => {
      if (itemName !== "") trigger(`${name}.name`);
      trigger(`${name}.type`);
    }, 25);
  }, [itemName]);

  const header = useMemo(
    () => `${name.split(".").slice(-1)}: (${itemType}) ${itemName}`,
    [name, itemName, itemType]
  );

  const convertedDefaults = useMemo(
    () => ({
      allow_invalid_certs:
        main?.allow_invalid_certs ??
        defaults?.allow_invalid_certs ??
        hard_defaults?.allow_invalid_certs,
      custom_headers: firstNonEmpty(
        main?.custom_headers,
        defaults?.custom_headers,
        hard_defaults?.custom_headers
      ),
      delay: firstNonDefault(
        main?.delay,
        defaults?.delay,
        hard_defaults?.delay
      ),
      desired_status_code: firstNonDefault(
        main?.desired_status_code,
        defaults?.desired_status_code,
        hard_defaults?.desired_status_code
      ),
      max_tries: firstNonDefault(
        main?.max_tries,
        defaults?.max_tries,
        hard_defaults?.max_tries
      ),
      secret: firstNonDefault(
        main?.secret,
        defaults?.secret,
        hard_defaults?.secret
      ),
      silent_fails:
        main?.silent_fails ??
        defaults?.silent_fails ??
        hard_defaults?.silent_fails,
      type: firstNonDefault(main?.type, defaults?.type, hard_defaults?.type),
      url: firstNonDefault(main?.url, defaults?.url, hard_defaults?.url),
    }),
    [main, defaults, hard_defaults]
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
                  mains &&
                  itemName !== "" &&
                  Object.keys(mains).indexOf(itemName) !== -1
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
            tooltip="Style of WebHook to emulate"
            options={webHookTypeOptions}
            position="right"
          />
          <FormItem
            name={`${name}.name`}
            required
            unique
            col_sm={12}
            label={"Name"}
          />
          <FormItem
            name={`${name}.url`}
            required
            col_sm={12}
            type="text"
            label="Target URL"
            tooltip="Where to send the WebHook"
            defaultVal={convertedDefaults.url}
            isURL
          />
          <BooleanWithDefault
            name={`${name}.allow_invalid_certs`}
            label="Allow Invalid Certs"
            defaultValue={convertedDefaults.allow_invalid_certs}
          />
          <FormItem
            name={`${name}.secret`}
            required
            col_sm={12}
            label="Secret"
            defaultVal={convertedDefaults.secret}
          />
          <FormKeyValMap
            name={`${name}.custom_headers`}
            defaults={convertedDefaults.custom_headers}
          />
          <FormItem
            name={`${name}.desired_status_code`}
            col_xs={6}
            label="Desired Status Code"
            tooltip="Treat the WebHook as successful when this status code is received (0=2XX)"
            isNumber
            defaultVal={convertedDefaults.desired_status_code}
          />
          <FormItem
            name={`${name}.max_tries`}
            col_xs={6}
            label="Max tries"
            isNumber
            defaultVal={convertedDefaults.max_tries}
            position="right"
          />
          <FormItem
            name={`${name}.delay`}
            col_sm={12}
            label="Delay"
            tooltip="Delay sending by this duration"
            defaultVal={convertedDefaults.delay}
            position="right"
          />
          <BooleanWithDefault
            name={`${name}.silent_fails`}
            label="Silent fails"
            tooltip="Notify if WebHook fails max tries times"
            defaultValue={convertedDefaults.silent_fails}
          />
        </Row>
      </Accordion.Body>
    </Accordion>
  );
};

export default EditServiceWebHook;
