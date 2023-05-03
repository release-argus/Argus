import { Button, ButtonGroup, Col } from "react-bootstrap";
import { Controller, useFormContext } from "react-hook-form";
import { FC, memo } from "react";
import {
  faCheckCircle,
  faCircleXmark,
} from "@fortawesome/free-solid-svg-icons";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { HelpTooltip } from "components/generic";

interface Props {
  name: string;

  col_xs?: number;
  col_sm?: number;
  label?: string;
  defaultValue?: boolean;
  tooltip?: string;
  onRight?: boolean;
}

const BooleanWithDefault: FC<Props> = ({
  name,
  col_xs = 12,
  col_sm = 12,
  label,
  defaultValue,
  tooltip,
  onRight,
}) => {
  const { control } = useFormContext();
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
    value: undefined,
    text: "Default: ",
    icon: defaultValue ? faCheckCircle : faCircleXmark,
    class: defaultValue ? "success" : "danger",
  };

  const leftPadding = [
    col_sm !== 12 && onRight ? "ps-2" : "",
    col_xs !== 12 && onRight ? "ps-2" : "",
  ].join(" ");
  const rightPadding = [
    col_sm !== 12 && !onRight ? "pe-2" : "",
    col_xs !== 12 && !onRight ? "pe-2" : "",
  ].join(" ");

  return (
    <Col
      xs={col_xs}
      sm={col_sm}
      className={`${leftPadding} ${rightPadding} pt-1 pb-1`}
      style={{ display: "flex", alignItems: "center" }}
    >
      <>
        {label && <a style={{ float: "left" }}>{label}</a>}
        {tooltip && <HelpTooltip text={tooltip} />}
      </>

      <div
        style={{ marginLeft: "auto", paddingLeft: "0.5rem", float: "right" }}
      >
        <Controller
          control={control}
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
                      className={`icon-${option.class}`}
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
                {`${optionsDefault.text} `}
                <FontAwesomeIcon
                  icon={optionsDefault.icon}
                  style={{
                    height: "1rem",
                  }}
                  className={`icon-${optionsDefault.class}`}
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
