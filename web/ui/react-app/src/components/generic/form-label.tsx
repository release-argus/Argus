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
    {required && <span className="icon-danger">*</span>}
    {tooltip && <HelpTooltip text={tooltip} />}
  </Form.Label>
);

export default FormLabel;
