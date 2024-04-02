import { Col, FormControl, FormGroup } from "react-bootstrap";
import { FC, useMemo } from "react";
import { useFormContext, useWatch } from "react-hook-form";

import FormLabel from "./form-label";
import { useError } from "hooks/errors";

interface Props {
  name: string;

  label: string;
  tooltip?: string;
  defaultVal?: string;
  placeholder?: string;
}

/**
 * Returns a form item with a preview image
 *
 * @param name - The name of the form item
 * @param label - The label of the form item
 * @param tooltip - The tooltip of the form item
 * @param defaultVal - The default value of the form item
 * @param placeholder - The placeholder of the form item
 * @returns A form item at name with a preview image, label and tooltip
 */
const FormItemWithPreview: FC<Props> = ({
  name,

  label,
  tooltip,
  defaultVal,
  placeholder,
}) => {
  const { register } = useFormContext();
  const formValue: string | undefined = useWatch({ name: name });
  const error = useError(name, true);
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
            style={{ marginRight: preview ? "1rem" : undefined }}
            autoFocus={false}
            {...register(name, {
              validate: (value: string | undefined) => {
                // Allow empty values
                if ((value ?? "") === "") return true;

                // Validate that it's a URL (with prefix)
                try {
                  const parsedURL = new URL(value as string);
                  if (!["http:", "https:"].includes(parsedURL.protocol))
                    throw new Error("Invalid protocol");
                } catch (error) {
                  if (/^https?:\/\//.test(value as string)) {
                    return "Invalid URL";
                  }
                  return "Invalid URL - http(s):// prefix required";
                }

                return true;
              },
            })}
            isInvalid={!!error}
          />
          {preview}
        </div>
      </FormGroup>
      {error && (
        <small className="error-msg">{error["message"] || "err"}</small>
      )}
    </Col>
  );
};

export default FormItemWithPreview;
