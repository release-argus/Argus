import { Col, FormControl, FormGroup } from "react-bootstrap";
import { FC, JSX, useMemo } from "react";

import FormLabel from "./form-label";
import { useFormContext } from "react-hook-form";

interface FormItemProps {
  name: string;
  required?: boolean;

  col_xs?: number;
  col_sm?: number;
  label?: string;
  tooltip?: string | JSX.Element;

  defaultVal?: string;
  placeholder?: string;

  rows?: number;
  onRight?: boolean;
  onMiddle?: boolean;
}

/**
 * Returns a form textarea
 *
 * @param name - The name of the form item
 * @param required - Whether the form item is required
 * @param col_xs - The number of columns the form item should take up on extra small screens
 * @param col_sm - The number of columns the form item should take up on small screens
 * @param label - The label of the form item
 * @param tooltip - The tooltip of the form item
 * @param defaultVal - The default value of the form item
 * @param placeholder - The placeholder of the form item
 * @param rows - The number of rows for the textarea
 * @param onRight - Whether the form item should be on the right
 * @param onMiddle - Whether the form item should be in the middle
 * @returns A form textarea with a label and tooltip
 */
const FormTextArea: FC<FormItemProps> = ({
  name,
  required,

  col_xs = 12,
  col_sm = 6,
  label,
  tooltip,

  defaultVal,
  placeholder,

  rows,
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
            validate: (value: string | undefined) => {
              // Validate that it's non-empty (including default value)
              if (required) {
                const testValue = value || defaultVal || "";
                const validation = /.+/.test(testValue);
                return validation ? true : "Required";
              }
              return true;
            },
          })}
        />
      </FormGroup>
    </Col>
  );
};

export default FormTextArea;
