import { Col, Form } from "react-bootstrap";
import { FC, useMemo } from "react";
import { useFormContext, useFormState } from "react-hook-form";

import FormLabel from "./form-label";
import { getNestedError } from "utils";

interface FormItemProps {
  name: string;
  registerParams?: Record<string, unknown>;
  required?: boolean;
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
  required,
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
      <Form.Group>
        {label && (
          <FormLabel
            text={label}
            tooltip={tooltip}
            required={required}
            small={!!smallLabel}
          />
        )}
        <Form.Control
          className={error && "form-error"}
          type={type}
          placeholder={defaultVal || placeholder}
          autoFocus={false}
          {...register(name, {
            validate: (value) => {
              let validation = true;
              const testValue = value || defaultVal || "";
              if (required) {
                validation = /.+/.test(testValue);
                if (!validation) return "Required";
              }

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
      </Form.Group>
    </Col>
  );
};

export default FormItem;
