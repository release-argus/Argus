import { Col, FormControl, FormGroup, InputGroup } from "react-bootstrap";
import { useFormContext, useWatch } from "react-hook-form";

import { FC } from "react";
import FormLabel from "./form-label";
import { formPadding } from "./util";
import { useError } from "hooks/errors";

interface FormItemColourProps {
  name: string;

  col_xs?: number;
  col_sm?: number;

  label: string;
  tooltip?: string;
  defaultVal?: string;
  position?: "left" | "middle" | "right";
  positionXS?: "left" | "middle" | "right";
}

/**
 * Returns a form item for a hex colour with a colour picker
 *
 * @param name - The name of the field
 * @param col_xs - The number of columns to take up on extra small screens
 * @param col_sm - The number of columns to take up on small screens
 * @param label - The form label to display
 * @param tooltip - The tooltip to display
 * @param defaultVal - The default value of the field
 * @param position - The position of the field
 * @param positionXS - The position of the field on extra small screens
 * @returns A form item for a hex colour with a colour picker, label and tooltip
 */
const FormItemColour: FC<FormItemColourProps> = ({
  name,

  col_xs = 12,
  col_sm = 6,

  label,
  tooltip,
  defaultVal,
  position = "left",
  positionXS = position,
}) => {
  const { register, setValue } = useFormContext();
  const hexColour: string = useWatch({ name: name });
  const trimmedHex = hexColour?.replace("#", "");
  const error = useError(name, true);
  const padding = formPadding({ col_xs, col_sm, position, positionXS });
  const setColour = (hex: string) =>
    setValue(name, hex.substring(1), { shouldDirty: true });

  return (
    <Col xs={col_xs} sm={col_sm} className={`${padding} pt-1 pb-1 col-form`}>
      <FormGroup style={{ display: "flex", flexDirection: "column" }}>
        <div>
          <FormLabel text={label} tooltip={tooltip} />
        </div>
        <div style={{ display: "flex", flexWrap: "nowrap" }}>
          <InputGroup className="mb-2">
            <InputGroup.Text>#</InputGroup.Text>
            <FormControl
              style={{ width: "25%" }}
              type="text"
              defaultValue={trimmedHex}
              placeholder={defaultVal}
              maxLength={6}
              autoFocus={false}
              {...register(name, {
                pattern: {
                  value: /^[\da-f]{6}$|^$/i,
                  message: "Invalid colour hex",
                },
              })}
              isInvalid={error !== undefined}
            />
            <FormControl
              className="form-control-color"
              style={{ width: "30%" }}
              type="color"
              title="Choose your color"
              value={`#${trimmedHex || defaultVal?.replace("#", "")}`}
              onChange={(event) => setColour(event.target.value)}
              autoFocus={false}
            />
          </InputGroup>
        </div>
      </FormGroup>
      {error && (
        <small className="error-msg">{error["message"] || "err"}</small>
      )}
    </Col>
  );
};

export default FormItemColour;
