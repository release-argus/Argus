import { Col, Form } from "react-bootstrap";
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
  placeholder?: string;
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
  placeholder,
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
    setValue(name, hex);
  };

  return (
    <Col xs={col_xs} sm={col_sm} className={`${padding} pt-1 pb-1 col-form`}>
      <Form.Group style={{ display: "flex", flexDirection: "column" }}>
        <div>
          <FormLabel text={label} tooltip={tooltip} />
        </div>
        <div style={{ display: "flex", flexWrap: "nowrap" }}>
          <Form.Control
            required={required}
            style={{ width: "50%" }}
            type="text"
            value={hexColour}
            placeholder={placeholder}
            autoFocus={false}
            {...register(name, {
              pattern: required ? /^#[\da-f]{6}$/ : /^#[\da-f]{6}$|^$/,
              required: required,
            })}
          />
          <Form.Control
            className="form-control-color"
            style={{ width: "50%" }}
            type="color"
            title="Choose your color"
            value={hexColour || placeholder}
            onChange={(event) => setColour(event.target.value)}
            autoFocus={false}
          />
        </div>
      </Form.Group>
    </Col>
  );
};

export default FormItemColour;
