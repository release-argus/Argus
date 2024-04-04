import { Button, ButtonGroup, Col } from "react-bootstrap";
import { Controller, useFormContext } from "react-hook-form";
import { FC, memo, useEffect } from "react";
import {
  faCheckCircle,
  faCircleXmark,
} from "@fortawesome/free-solid-svg-icons";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { HelpTooltip } from "components/generic";
import { strToBool } from "utils";

interface Props {
  name: string;

  label?: string;
  tooltip?: string;
  defaultValue?: boolean;
}

/**
 * Returns a form field with buttons to choose between true, false, and default
 *
 * @param name - The name of the field
 * @param label - The form label to display
 * @param tooltip - The tooltip to display
 * @param defaultValue - The default value of the field
 * @returns A form field at name with a label, tooltip and buttons to choose between true, false, and default
 */
const BooleanWithDefault: FC<Props> = ({
  name,
  label,
  defaultValue,
  tooltip,
}) => {
  const { getValues, setValue } = useFormContext();
  const options = [
    {
      value: true,
      icon: faCheckCircle,
      class: "success",
      text: "Yes",
    },
    {
      value: false,
      icon: faCircleXmark,
      class: "danger",
      text: "No",
    },
  ];
  const optionsDefault = {
    value: null,
    text: "Default: ",
    icon: defaultValue ? faCheckCircle : faCircleXmark,
    class: defaultValue ? "success" : "danger",
  };
  // on load, convert the value if it's a string
  useEffect(() => {
    setValue(name, strToBool(getValues(name)));
  }, []);

  return (
    <Col
      xs={12}
      className="pt-1 pb-1"
      style={{ display: "flex", alignItems: "center" }}
    >
      <>
        {label && <a style={{ float: "left" }}>{label}</a>}
        {tooltip && <HelpTooltip text={tooltip} />}
      </>

      <div
        style={{
          display: "flex",
          flexWrap: "wrap",
          justifyContent: "flex-end",
          marginLeft: "auto",
          paddingLeft: "0.5rem",
        }}
      >
        <Controller
          name={name}
          render={({ field: { onChange, value } }) => (
            <>
              <ButtonGroup>
                {options.map((option) => (
                  <Button
                    name={`${name}-${option.value}`}
                    key={option.class}
                    id={`option-${option.value}`}
                    className={`btn-${
                      value === option.value ? "" : "un"
                    }checked pad-no`}
                    onClick={() => onChange(option.value)}
                    variant="secondary"
                  >
                    {`${option.text} `}
                    <FontAwesomeIcon
                      icon={option.icon}
                      style={{
                        height: "1rem",
                      }}
                      className={`text-${option.class}`}
                    />
                  </Button>
                ))}
              </ButtonGroup>
              <>{"  |  "}</>
              <Button
                name={`${name}-${optionsDefault.value}`}
                id={`option-${optionsDefault.value}`}
                className={`btn-${
                  value === optionsDefault.value ? "" : "un"
                }checked pad-no`}
                onClick={() => onChange(optionsDefault.value)}
                variant="secondary"
              >
                {optionsDefault.text}
                <FontAwesomeIcon
                  icon={optionsDefault.icon}
                  style={{
                    height: "1rem",
                  }}
                  className={`text-${optionsDefault.class}`}
                />
              </Button>
            </>
          )}
        />
      </div>
    </Col>
  );
};

export default memo(BooleanWithDefault);
