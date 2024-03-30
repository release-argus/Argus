import { Col, FormCheck as FormCheckRB, FormGroup } from "react-bootstrap";
import { FC, JSX, useMemo } from "react";

import { FormCheckType } from "react-bootstrap/esm/FormCheck";
import FormLabel from "./form-label";
import { useFormContext } from "react-hook-form";

interface FormCheckProps {
  name: string;

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

/**
 * Returns a form checkbox
 *
 * @param name - The name of the field
 * @param col_xs - The number of columns to take up on extra small screens
 * @param col_sm - The number of columns to take up on small screens
 * @param size - The size of the checkbox
 * @param label - The form label to display
 * @param smallLabel - Whether the label should be small
 * @param tooltip - The tooltip to display
 * @param type - The type of the checkbox
 * @param onRight - Whether the checkbox should be on the right
 * @param onMiddle - Whether the checkbox should be in the middle
 * @returns A form checkbox with a label and tooltip
 */
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
          {...register(name)}
        />
      </FormGroup>
    </Col>
  );
};

export default FormCheck;
