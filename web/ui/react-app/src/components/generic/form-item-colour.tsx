import { Col, FormControl, FormGroup } from "react-bootstrap";
import { FC, useMemo, useState } from "react";

import FormLabel from "./form-label";
import { useFormContext } from "react-hook-form";

interface FormItemColourProps {
  name: string;
  required?: boolean;

  col_xs?: number;
  col_sm?: number;
  label: string;
  tooltip?: string;
  rows?: number;
  options?: JSX.Element[];
  value?: string;
  defaultVal?: string;
  onRight?: boolean;
  onMiddle?: boolean;
}

const FormItemColour: FC<FormItemColourProps> = ({
  name,
  required,

  col_xs = 12,
  col_sm = 6,

  label,
  tooltip,
  value,
  defaultVal,
  onRight,
  onMiddle,
}) => {
  const { register, setValue } = useFormContext();
  const padding = useMemo(() => {
    return [
      col_sm !== 12 && onRight ? "ps-sm-2" : "",
      col_xs !== 12 && onRight ? "ps-2" : "",
      col_sm !== 12 && !onRight ? (onMiddle ? "ps-sm-2" : "pe-sm-2") : "",
      col_xs !== 12 && !onRight ? (onMiddle ? "ps-2" : "pe-2") : "",
    ].join(" ");
  }, []);
  const [hexColour, setHexColour] = useState(value);
  const setColour = (hex: string) => {
    setHexColour(hex);
    setValue(name, hex, { shouldDirty: true });
  };

  return (
    <Col xs={col_xs} sm={col_sm} className={`${padding} pt-1 pb-1 col-form`}>
      <FormGroup style={{ display: "flex", flexDirection: "column" }}>
        <div>
          <FormLabel text={label} tooltip={tooltip} />
        </div>
        <div style={{ display: "flex", flexWrap: "nowrap" }}>
          <FormControl
            required={required}
            style={{ width: "50%" }}
            type="text"
            value={hexColour}
            placeholder={defaultVal}
            autoFocus={false}
            {...register(name, {
              pattern: required ? /^#[\da-f]{6}$/ : /^#[\da-f]{6}$|^$/,
              required: required,
            })}
          />
          <FormControl
            className="form-control-color"
            style={{ width: "50%" }}
            type="color"
            title="Choose your color"
            value={hexColour || defaultVal}
            onChange={(event) => setColour(event.target.value)}
            autoFocus={false}
          />
        </div>
      </FormGroup>
    </Col>
  );
};

export default FormItemColour;
