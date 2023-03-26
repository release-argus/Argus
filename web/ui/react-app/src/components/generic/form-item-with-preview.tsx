import { Col, Form } from "react-bootstrap";
import { FC, useMemo } from "react";
import { useFormContext, useWatch } from "react-hook-form";

import FormLabel from "./form-label";

interface FormItemWithPreviewProps {
  name: string;

  label: string;
  tooltip?: string;
  placeholder?: string;
}

const FormItemWithPreview: FC<FormItemWithPreviewProps> = ({
  name,

  label,
  tooltip,
  placeholder,
}) => {
  const { register } = useFormContext();
  const formValue = useWatch({ name: name });
  const preview = useMemo(() => {
    const url = formValue || placeholder || "";
    try {
      new URL(url);
      // Render the image if it's a valid URL that resolved
      return (
        <div
          style={{ maxWidth: "100%", overflow: "hidden", marginLeft: "auto" }}
        >
          <img
            src={url}
            alt="Icon preview"
            style={{ height: "2em", width: "auto" }}
          />
        </div>
      );
    } catch (error) {
      return false;
    }
  }, [formValue, placeholder]);

  return (
    <Col xs={12} sm={12} className={"pt-1 pb-1 col-form"}>
      <Form.Group>
        <FormLabel text={label} tooltip={tooltip} />
        <div style={{ display: "flex", alignItems: "center" }}>
          <Form.Control
            type="text"
            value={formValue}
            placeholder={placeholder}
            pattern="^https?://.+"
            style={{ width: preview ? "90%" : "100%" }}
            autoFocus={false}
            {...register(name)}
          />
          {preview}
        </div>
      </Form.Group>
    </Col>
  );
};

export default FormItemWithPreview;
