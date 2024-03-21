import { Col, Form, FormGroup } from "react-bootstrap";
import { Controller, useFormState } from "react-hook-form";
import { FC, JSX, useMemo } from "react";

import FormLabel from "./form-label";
import { OptionType } from "types/util";
import { getNestedError } from "utils";

interface FormSelectProps {
  name: string;
  required?: boolean;
  customValidation?: (value: string) => string | boolean;

  key?: string;
  col_xs?: number;
  col_sm?: number;
  label?: string;
  smallLabel?: boolean;
  tooltip?: string | JSX.Element;

  options: OptionType[];

  isURL?: boolean;

  onRight?: boolean;
  onRightXS?: boolean;
  onMiddle?: boolean;
}

const FormSelect: FC<FormSelectProps> = ({
  name,
  required,
  customValidation,

  key = name,
  col_xs = 12,
  col_sm = 6,
  label,
  smallLabel,
  tooltip,
  options,
  onRight,
  onRightXS,
  onMiddle,
}) => {
  const { errors } = useFormState();
  const error = customValidation && getNestedError(errors, name);

  const padding = useMemo(() => {
    const paddingClasses = [];

    // Padding for being on the right
    if (onRight) {
      if (col_sm !== 12) paddingClasses.push("ps-sm-2");
    }
    if (onRight || onRightXS) {
      if (col_xs !== 12) paddingClasses.push("ps-xs-2");
    }
    // Padding for being in the middle
    else if (onMiddle) {
      if (col_sm !== 12) {
        paddingClasses.push("ps-sm-1");
        paddingClasses.push("pe-sm-1");
      }
      if (!onRightXS && col_xs !== 12) {
        paddingClasses.push("ps-xs-1");
        paddingClasses.push("pe-xs-1");
      }
    }
    // Padding for being on the left
    else {
      if (col_sm !== 12) paddingClasses.push("pe-sm-2");
      if (col_xs !== 12) paddingClasses.push("pe-2");
    }

    return paddingClasses.join(" ");
  }, [col_xs, col_sm, onRight, onMiddle]);

  return (
    <Col
      xs={col_xs}
      sm={col_sm}
      className={`${padding} pt-1 pb-1 col-form`}
      key={key}
    >
      <FormGroup>
        {label && (
          <FormLabel text={label} tooltip={tooltip} small={!!smallLabel} />
        )}
        <Controller
          name={name}
          render={({ field }) => (
            <Form.Select {...field} aria-label={label} required={required}>
              {options.map((opt) => (
                <option
                  className="form-select-option"
                  value={opt.value}
                  key={opt.label}
                >
                  {opt.label}
                </option>
              ))}
            </Form.Select>
          )}
          rules={{
            validate: (value) => {
              if (customValidation) return customValidation(value);
            },
          }}
        />
        {error && (
          <small className="error-msg">{error["message"] || "err"}</small>
        )}
      </FormGroup>
    </Col>
  );
};

export default FormSelect;
