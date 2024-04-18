import { Col, Form, FormGroup } from "react-bootstrap";
import { Controller, useFormContext } from "react-hook-form";
import { FC, JSX } from "react";

import FormLabel from "./form-label";
import { OptionType } from "types/util";
import { Position } from "types/config";
import { formPadding } from "./util";
import { useError } from "hooks/errors";

interface FormSelectProps {
  name: string;
  customValidation?: (value: string) => string | boolean;
  onChange?: (e: React.ChangeEvent<HTMLSelectElement>) => void;

  key?: string;
  col_xs?: number;
  col_sm?: number;
  label?: string;
  smallLabel?: boolean;
  tooltip?: string | JSX.Element;

  options: OptionType[];

  position?: Position;
  positionXS?: Position;
}

/**
 * FormSelect is a labelled select form item
 *
 * @param name - The name of the form item
 * @param required - Whether the form item is required
 * @param customValidation - Custom validation function for the form item
 * @param onChange - The function to call when the form item changes
 * @param key - The key of the form item
 * @param col_xs - The number of columns the item takes up on XS+ screens
 * @param col_sm - The number of columns the item takes up on SM+ screens
 * @param label - The label of the form item
 * @param smallLabel - Whether the label should be small
 * @param tooltip - The tooltip of the form item
 * @param options - The options for the select field
 * @param position - The position of the form item
 * @param positionXS - The position of the form item on extra small screens
 * @returns A labeled select form item
 */
const FormSelect: FC<FormSelectProps> = ({
  name,
  customValidation,
  onChange,

  key = name,
  col_xs = 12,
  col_sm = 6,
  label,
  smallLabel,
  tooltip,
  options,
  position = "left",
  positionXS = position,
}) => {
  const { setValue } = useFormContext();
  const error = useError(name, customValidation !== undefined);
  const padding = formPadding({ col_xs, col_sm, position, positionXS });

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
            <Form.Select
              {...field}
              aria-label={label}
              onChange={onChange || ((e) => setValue(name, e.target.value))}
            >
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
            validate: {
              customValidation: (value) =>
                customValidation ? customValidation(value) : undefined,
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
