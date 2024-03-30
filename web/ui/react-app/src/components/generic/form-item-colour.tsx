import { Col, FormControl, FormGroup } from "react-bootstrap";
import { FC, useMemo, useState } from "react";

import FormLabel from "./form-label";
import { useFormContext } from "react-hook-form";

interface FormItemColourProps {
  name: string;

  col_xs?: number;
  col_sm?: number;

  label: string;
  tooltip?: string;
  value?: string;
  defaultVal?: string;
  onRight?: boolean;
  onMiddle?: boolean;
}

/**
 * Returns a form item for a hex colour with a colour picker
 *
 * @param name - The name of the field
 * @param col_xs - The number of columns to take up on extra small screens
 * @param col_sm - The number of columns to take up on small screens
 * @param label - The form label to display
 * @param tooltip - The tooltip to display
 * @param value - The value of the field
 * @param defaultVal - The default value of the field
 * @param onRight - Whether the form item should be on the right
 * @param onMiddle - Whether the form item should be in the middle
 * @returns A form item for a hex colour with a colour picker, label and tooltip
 */
const FormItemColour: FC<FormItemColourProps> = ({
  name,

  col_xs = 12,
  col_sm = 6,

  label,
  tooltip,
  value,
  defaultVal,
  onRight,
  onMiddle,
}) => {
  const { register, setValue } = useFormContext();
  const padding = useMemo(() => {
    return [
      col_sm !== 12 && onRight ? "ps-sm-2" : "",
      col_xs !== 12 && onRight ? "ps-2" : "",
      col_sm !== 12 && !onRight ? (onMiddle ? "ps-sm-2" : "pe-sm-2") : "",
      col_xs !== 12 && !onRight ? (onMiddle ? "ps-2" : "pe-2") : "",
    ].join(" ");
  }, []);
  const [hexColour, setHexColour] = useState(value);
  const setColour = (hex: string) => {
    setHexColour(hex);
    setValue(name, hex, { shouldDirty: true });
  };

  return (
    <Col xs={col_xs} sm={col_sm} className={`${padding} pt-1 pb-1 col-form`}>
      <FormGroup style={{ display: "flex", flexDirection: "column" }}>
        <div>
          <FormLabel text={label} tooltip={tooltip} />
        </div>
        <div style={{ display: "flex", flexWrap: "nowrap" }}>
          <FormControl
            style={{ width: "50%" }}
            type="text"
            value={hexColour}
            placeholder={defaultVal}
            autoFocus={false}
            {...register(name, {
              pattern: {
                value: /^[\da-f]{6}$|^$/i,
                message: "Invalid colour hex",
              },
            })}
          />
          <FormControl
            className="form-control-color"
            style={{ width: "50%" }}
            type="color"
            title="Choose your color"
            value={hexColour || defaultVal}
            onChange={(event) => setColour(event.target.value)}
            autoFocus={false}
          />
        </div>
      </FormGroup>
    </Col>
  );
};

export default FormItemColour;
