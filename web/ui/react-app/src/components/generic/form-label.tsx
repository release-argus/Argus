import { FC, JSX } from "react";

import { Form } from "react-bootstrap";
import { HelpTooltip } from "components/generic";

interface FormLabelProps {
  text: string;
  tooltip?: string | JSX.Element;
  heading?: boolean;
  required?: boolean;
  small?: boolean;
}

/**
 * Returns a label for a form item
 *
 * @param text - The text of the label
 * @param tooltip - The tooltip of the label
 * @param heading - Whether the label is a heading
 * @param required - Whether the label is required
 * @param small - Whether the label is small
 * @returns A label for a form item
 */
const FormLabel: FC<FormLabelProps> = ({
  text,
  tooltip,
  heading,
  required,
  small,
}: FormLabelProps) => (
  <Form.Label
    style={
      heading
        ? {
            fontSize: "1.25rem",
            textDecorationLine: "underline",
            paddingTop: "1.5rem",
          }
        : small
        ? { fontSize: "0.8rem" }
        : {}
    }
  >
    {text}
    {required && <span className="text-danger">*</span>}
    {tooltip && <HelpTooltip text={tooltip} />}
  </Form.Label>
);

export default FormLabel;
