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

  isURL?: boolean;
  isRegex?: boolean;
  defaultVal?: string;
  placeholder?: string;

  onRight?: boolean;
  onMiddle?: boolean;
}

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
  isURL,
  isRegex,
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
    const paddingClasses = [];

    // Padding for being on the right
    if (onRight) {
      if (col_sm !== 12) paddingClasses.push("ps-sm-2");
      if (col_xs !== 12) paddingClasses.push("ps-2");
    }
    // Padding for being in the middle
    else if (onMiddle) {
      if (col_sm !== 12) {
        paddingClasses.push("ps-sm-1");
        paddingClasses.push("pe-sm-1");
      }
      if (col_xs !== 12) {
        paddingClasses.push("ps-1");
        paddingClasses.push("pe-1");
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
            validate: (value) => {
              let validation = true;
              const testValue = value || defaultVal || "";
              if (required) validation = /.+/.test(testValue);
              if (!validation) return required === true ? "Required" : required;

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

              if (isRegex) {
                try {
                  new RegExp(testValue);
                } catch (error) {
                  return "Invalid RegEx";
                }
              }

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
                return validation || "Should be unique";
              }

              return validation || "error";
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
