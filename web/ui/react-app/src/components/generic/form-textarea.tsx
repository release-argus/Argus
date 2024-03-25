import { Col, FormControl, FormGroup } from "react-bootstrap";
import { FC, JSX } from "react";

import FormLabel from "./form-label";
import { formPadding } from "./util";
import { useError } from "hooks/errors";
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
  position?: "left" | "middle" | "right";
  positionXS?: "left" | "middle" | "right";
}

/**
 * FormTextArea is a labelled form textarea
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
 * @param position - The position of the form item
 * @param positionXS - The position of the form item on extra small screens
 * @returns  A labeled form textarea
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
  position = "left",
  positionXS = position,
}) => {
  const { register } = useFormContext();
  const error = useError(name, required);

  const padding = formPadding({ col_xs, col_sm, position, positionXS });

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
              const testValue = value || defaultVal || "";
              // Validate that it's non-empty (including default value)
              if (required) {
                const validation = /.+/.test(testValue);
                return validation ? true : "Required";
              }
              return true;
            },
          })}
          isInvalid={!!error}
        />
        {error && (
          <small className="error-msg">{error["message"] || "err"}</small>
        )}
      </FormGroup>
    </Col>
  );
};

export default FormTextArea;
