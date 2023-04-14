import { Button, ButtonGroup, Col } from "react-bootstrap";
import { FC, memo, useEffect, useState } from "react";
import {
  faCheckCircle,
  faCircleXmark,
} from "@fortawesome/free-solid-svg-icons";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { HelpTooltip } from "components/generic";
import { useFormContext } from "react-hook-form";

interface Props {
  name: string;

  col_xs?: number;
  col_sm?: number;
  label?: string;
  value?: boolean;
  defaultValue?: boolean;
  tooltip?: string;
  onRight?: boolean;
}

const BooleanWithDefault: FC<Props> = ({
  name,
  col_xs = 12,
  col_sm = 12,
  label,
  value,
  defaultValue,
  tooltip,
  onRight,
}) => {
  const { getValues, setValue } = useFormContext();
  const [bool, setBoolState] = useState<boolean | undefined>(value);
  useEffect(() => {
    const val = getValues(name);
    // have no starting value, but have a value in the form
    if (value === undefined && val !== undefined) {
      setBoolState(typeof val === "boolean" ? val : val === "true");
    }
  }, []);
  const setBool = (newVal?: boolean) => {
    setBoolState(newVal);
    setValue(name, newVal);
  };
  const options = [
    {
      value: true,
      icon: faCheckCircle,
      class: "success",
      checked: bool === true,
      text: "Yes",
    },
    {
      value: false,
      icon: faCircleXmark,
      class: "danger",
      checked: bool === false,
      text: "No",
    },
  ];
  const optionsDefault = {
    value: undefined,
    text: "Default: ",
    checked: bool === undefined,
    ...(defaultValue
      ? {
          icon: faCheckCircle,
          class: "success",
        }
      : {
          icon: faCircleXmark,
          class: "danger",
        }),
  };
  const optionsButtons = (
    <ButtonGroup>
      {options.map((option) => (
        <Button
          name={`${name}-${option.value}`}
          key={option.class}
          id={`option-${option.value}`}
          className={`btn-${option.checked ? "" : "un"}checked pad-no`}
          onClick={() => setBool(option.value)}
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
  );
  const defaultButton = (
    <Button
      name={`${name}-${optionsDefault.value}`}
      id={`option-${optionsDefault.value}`}
      className={`btn-${optionsDefault.checked ? "" : "un"}checked pad-no`}
      onClick={() => setBool(optionsDefault.value)}
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
  );

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
        {optionsButtons}
        <>{"  |  "}</>
        {defaultButton}
      </div>
    </Col>
  );
};

export default memo(BooleanWithDefault);
