import { Col, FormControl, FormGroup } from "react-bootstrap";
import { FC, JSX, useMemo } from "react";

import FormLabel from "./form-label";
import { useFormContext } from "react-hook-form";

interface FormItemProps {
  name: string;
  registerParams?: Record<string, unknown>;
  required?: boolean;

  col_xs?: number;
  col_sm?: number;
  label?: string;
  tooltip?: string | JSX.Element;

  rows?: number;

  value?: string | number;

  isURL?: boolean;
  defaultVal?: string;
  placeholder?: string;

  onRight?: boolean;
  onMiddle?: boolean;
}

const FormTextArea: FC<FormItemProps> = ({
  name,
  registerParams = {},
  required,

  col_xs = 12,
  col_sm = 6,
  label,
  tooltip,
  rows,
  defaultVal,
  placeholder,
  onRight,
  onMiddle,
}) => {
  const { register } = useFormContext();
  const padding = useMemo(() => {
    return [
      col_sm !== 12 && onRight ? "ps-sm-2" : "",
      col_xs !== 12 && onRight ? "ps-2" : "",
      col_sm !== 12 && !onRight ? (onMiddle ? "ps-sm-2" : "pe-sm-2") : "",
      col_xs !== 12 && !onRight ? (onMiddle ? "ps-2" : "pe-2") : "",
    ].join(" ");
  }, [col_xs, col_sm, onRight, onMiddle]);
  return (
    <Col xs={col_xs} sm={col_sm} className={`${padding} pt-1 pb-1 col-form`}>
      <FormGroup>
        {label && (
          <FormLabel text={label} tooltip={tooltip} required={required} />
        )}
        <FormControl
          type={"textarea"}
          as="textarea"
          rows={rows}
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

              return validation || "error";
            },
            ...registerParams,
          })}
        />
      </FormGroup>
    </Col>
  );
};

export default FormTextArea;
