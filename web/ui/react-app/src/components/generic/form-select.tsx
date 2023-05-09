import { Col, Form } from "react-bootstrap";
import { Controller, useFormState } from "react-hook-form";
import { FC, useMemo } from "react";

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
  onMiddle,
}) => {
  const { errors } = useFormState();
  const error = customValidation && getNestedError(errors, name);

  const padding = useMemo(() => {
    return [
      col_sm !== 12 && onRight ? "ps-sm-2" : "",
      col_xs !== 12 && onRight ? "ps-2" : "",
      col_sm !== 12 && !onRight
        ? onMiddle
          ? "ps-sm-1 pe-sm-1"
          : "pe-sm-2"
        : "",
      col_xs !== 12 && !onRight ? (onMiddle ? "ps-2 pe-2" : "pe-2") : "",
    ].join(" ");
  }, []);
  return (
    <Col
      xs={col_xs}
      sm={col_sm}
      className={`${padding} pt-1 pb-1 col-form`}
      key={key}
    >
      <Form.Group>
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
              if (customValidation) {
                return customValidation(value);
              }
            },
          }}
        />
        {error && (
          <small className="error-msg">{error["message"] || "err"}</small>
        )}
      </Form.Group>
    </Col>
  );
};

export default FormSelect;
