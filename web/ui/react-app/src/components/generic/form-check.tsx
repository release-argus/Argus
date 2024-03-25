import { Col, FormCheck as FormCheckRB, FormGroup } from "react-bootstrap";
import { FC, JSX, useMemo } from "react";

import { FormCheckType } from "react-bootstrap/esm/FormCheck";
import FormLabel from "./form-label";
import { formPadding } from "./util";
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

  position?: "left" | "middle" | "right";
  positionXS?: "left" | "middle" | "right";
}

/**
 * FormCheck is labelled form check
 *
 * @param name - The name of the field
 * @param col_xs - The number of columns to take up on extra small screens
 * @param col_sm - The number of columns to take up on small screens
 * @param size - The size of the form check
 * @param label - The form label to display
 * @param smallLabel - Whether the label should be small
 * @param tooltip - The tooltip to display
 * @param type - The type of the form check (checkbox/radio/switch)
 * @param position - The position of the form check
 * @param positionXS - The position of the form check on extra small screens
 * @returns A labaled form check
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

  position = "left",
  positionXS = position,
}) => {
  const { register } = useFormContext();
  const padding = formPadding({ col_xs, col_sm, position, positionXS });

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
