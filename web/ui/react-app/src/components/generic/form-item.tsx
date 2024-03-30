import { Col, FormControl, FormGroup } from "react-bootstrap";
import { FC, JSX, useMemo } from "react";
import { useFormContext, useFormState } from "react-hook-form";

import FormLabel from "./form-label";
import { getNestedError } from "utils";

interface FormItemProps {
  name: string;
  registerParams?: Record<string, unknown>;
  required?: boolean | string;
  unique?: boolean;

  col_xs?: number;
  col_sm?: number;
  label?: string;
  smallLabel?: boolean;
  tooltip?: string | JSX.Element;
  type?: "text" | "number" | "url";

  isRegex?: boolean;
  isURL?: boolean;
  defaultVal?: string;
  placeholder?: string;

  onRight?: boolean;
  onMiddle?: boolean;
}

/**
 * Returns a form item
 *
 * @param name - The name of the form item
 * @param registerParams - Additional parameters for the form item
 * @param required - Whether the form item is required
 * @param unique - Whether the form item should be unique
 * @param col_xs - The number of columns the form item should take up on extra small screens
 * @param col_sm - The number of columns the form item should take up on small screens
 * @param label - The label of the form item
 * @param smallLabel - Whether the label should be small
 * @param tooltip - The tooltip of the form item
 * @param type - The type of the form item
 * @param isRegex - Whether the form item should be a regex
 * @param isURL - Whether the form item should be a URL
 * @param defaultVal - The default value of the form item
 * @param placeholder - The placeholder of the form item
 * @param onRight - Whether the form item should be on the right
 * @param onMiddle - Whether the form item should be in the middle
 * @returns A form item at name with a label and tooltip
 */
const FormItem: FC<FormItemProps> = ({
  name,
  registerParams = {},
  required = false,
  unique,

  col_xs = 12,
  col_sm = 6,
  label,
  smallLabel,
  tooltip,
  type = "text",
  isRegex,
  isURL,
  defaultVal,
  placeholder,

  onRight,
  onMiddle,
}) => {
  const { getValues, register } = useFormContext();
  const { errors } = useFormState();
  const error =
    (required || isURL || isRegex || registerParams["validate"]) &&
    getNestedError(errors, name);

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
          <FormLabel
            text={label}
            tooltip={tooltip}
            required={required !== false}
            small={!!smallLabel}
          />
        )}
        <FormControl
          className={error && "form-error"}
          type={type}
          placeholder={defaultVal || placeholder}
          autoFocus={false}
          {...register(name, {
            validate: (value: string | undefined) => {
              let validation = true;
              const testValue = value || defaultVal || "";

              // Validate that it's non-empty (including default value)
              if (required) {
                validation = /.+/.test(testValue);
                if (!validation)
                  return required === true ? "Required" : required;
              }

              // Validate that it's valid RegEx
              if (isRegex) {
                try {
                  new RegExp(testValue);
                } catch (error) {
                  return "Invalid RegEx";
                }
              }

              // Validate that it's a URL (with prefix)
              if (isURL) {
                try {
                  validation = required
                    ? new URL(testValue).protocol.startsWith("http")
                    : true;
                  if (!validation)
                    return "Invalid URL - http(s):// prefix required";
                } catch (error) {
                  return "Invalid URL";
                }
              }

              // Should be unique if it's changed from the default
              if (unique && testValue !== defaultVal) {
                const parts = name.split(".");
                const parent = parts.slice(0, parts.length - 2).join(".");
                const values = getValues(parent);
                const uniqueName = parts[parts.length - 1];
                validation =
                  value === ""
                    ? false
                    : values &&
                      values
                        .map(
                          (item: { [x: string]: string }) => item[uniqueName]
                        )
                        .filter((item: string) => item === value).length === 1;
                return validation || "Must be unique";
              }

              return validation;
            },
            ...registerParams,
          })}
        />
        {error && (
          <small className="error-msg">{error["message"] || "err"}</small>
        )}
      </FormGroup>
    </Col>
  );
};

export default FormItem;
