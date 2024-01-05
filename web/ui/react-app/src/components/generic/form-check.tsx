import { Col, FormCheck as FormCheckRB, FormGroup } from "react-bootstrap";
import { FC, JSX, useMemo } from "react";

import { FormCheckType } from "react-bootstrap/esm/FormCheck";
import FormLabel from "./form-label";
import { useFormContext } from "react-hook-form";

interface FormCheckProps {
  name?: string;

  col_xs?: number;
  col_sm?: number;
  size?: "sm" | "lg";
  label?: string;
  smallLabel?: boolean;
  tooltip?: string | JSX.Element;
  type?: FormCheckType;

  onRight?: boolean;
  onMiddle?: boolean;
}

const FormCheck: FC<FormCheckProps> = ({
  name,

  col_xs = 12,
  col_sm = 6,
  size = "sm",
  label,
  smallLabel,
  tooltip,
  type = "checkbox",

  onRight,
  onMiddle,
}) => {
  const { register } = useFormContext();

  const padding = useMemo(() => {
    return [
      col_sm !== 12
        ? onRight
          ? "ps-sm-2"
          : onMiddle
          ? "ps-sm-1 pe-sm-1"
          : "pe-sm-2"
        : "",
      col_xs !== 12 ? (onRight ? "ps-2" : onMiddle ? "ps-1 pe-1" : "pe-2") : "",
    ].join(" ");
  }, [col_xs, col_sm, onRight, onMiddle]);

  const registrationProps = useMemo(() => {
    return name ? { ...register(name) } : {};
  }, [name]);

  return (
    <Col xs={col_xs} sm={col_sm} className={`${padding} pt-1 pb-1 col-form`}>
      <FormGroup>
        {label && (
          <FormLabel text={label} tooltip={tooltip} small={!!smallLabel} />
        )}
        <FormCheckRB
          className={`form-check${size === "lg" ? "-large" : ""}`}
          type={type}
          autoFocus={false}
          {...registrationProps}
        />
      </FormGroup>
    </Col>
  );
};

export default FormCheck;
