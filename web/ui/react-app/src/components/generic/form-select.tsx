import { Col, Form, FormGroup } from "react-bootstrap";
import { Controller, useFormState } from "react-hook-form";
import { FC, JSX, useMemo } from "react";

import FormLabel from "./form-label";
import { OptionType } from "types/util";
import { getNestedError } from "utils";

interface FormSelectProps {
  name: string;
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

/**
 * FormSelect is a labelled select form item
 *
 * @param name - The name of the form item
 * @param required - Whether the form item is required
 * @param customValidation - Custom validation function for the form item
 * @param key - The key of the form item
 * @param col_xs - The number of columns the form item should take up on extra small screens
 * @param col_sm - The number of columns the form item should take up on small screens
 * @param label - The label of the form item
 * @param smallLabel - Whether the label should be small
 * @param tooltip - The tooltip of the form item
 * @param options - The options for the select field
 * @param onRight - Whether the form item should be on the right
 * @param onMiddle - Whether the form item should be in the middle
 * @returns A labeled select form item
 */
const FormSelect: FC<FormSelectProps> = ({
  name,
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
      <FormGroup>
        {label && (
          <FormLabel text={label} tooltip={tooltip} small={!!smallLabel} />
        )}
        <Controller
          name={name}
          render={({ field }) => (
            <Form.Select {...field} aria-label={label}>
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
