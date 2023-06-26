import { Col, FormControl, FormGroup } from "react-bootstrap";
import { FC, useMemo } from "react";
import { useFormContext, useWatch } from "react-hook-form";

import FormLabel from "./form-label";

interface Props {
  name: string;

  label: string;
  tooltip?: string;
  defaultVal?: string;
  placeholder?: string;
}

const FormItemWithPreview: FC<Props> = ({
  name,

  label,
  tooltip,
  defaultVal,
  placeholder,
}) => {
  const { register } = useFormContext();
  const formValue = useWatch({ name: name });
  const preview = useMemo(() => {
    const url = formValue || defaultVal || "";
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
  }, [formValue, defaultVal]);

  return (
    <Col xs={12} sm={12} className={"pt-1 pb-1 col-form"}>
      <FormGroup>
        <FormLabel text={label} tooltip={tooltip} />
        <div style={{ display: "flex", alignItems: "center" }}>
          <FormControl
            type="text"
            value={formValue}
            placeholder={placeholder || defaultVal}
            pattern="^https?://.+"
            style={{ width: preview ? "90%" : "100%" }}
            autoFocus={false}
            {...register(name)}
          />
          {preview}
        </div>
      </FormGroup>
    </Col>
  );
};

export default FormItemWithPreview;
