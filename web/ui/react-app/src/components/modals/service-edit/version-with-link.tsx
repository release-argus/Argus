import {
  Button,
  Col,
  FormControl,
  FormGroup,
  InputGroup,
} from "react-bootstrap";
import { FC, useEffect, useState } from "react";
import { useFormContext, useWatch } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormLabel } from "components/generic/form";
import { Position } from "types/config";
import { faLink } from "@fortawesome/free-solid-svg-icons";
import { formPadding } from "components/generic/util";
import { useError } from "hooks/errors";

interface Props {
  name: string;
  type: "github" | "url";
  required?: boolean;
  col_sm: number;
  col_xs: number;
  position?: Position;
}

/**
 * Returns the version field with a link
 *
 * @param name - The name of the field in the form
 * @param type - The type of version field
 * @param required - Whether the field is required
 * @param col_xs - The number of columns the item takes up on XS+ screens
 * @param col_sm - The number of columns the item takes up on SM+ screens
 * @param position - The position of the field
 * @returns The form fields for the `latest_version`
 */
const VersionWithLink: FC<Props> = ({
  name,
  type,
  required,
  col_xs = 12,
  col_sm = 6,
  position,
}) => {
  const { setValue, setError, clearErrors } = useFormContext();
  const value: string = useWatch({ name: name });

  const [isUnfocused, setIsUnfocused] = useState(true);
  const handleFocus = () => {
    setIsUnfocused(false);
  };
  const handleBlur = () => {
    setIsUnfocused(true);
  };
  const link = (type: "github" | "url") =>
    type === "github" ? `https://github.com/${value}` : value;

  const error = useError(name, required ? true : false);

  const padding = formPadding({ col_xs, col_sm, position });

  useEffect(() => {
    if (!isUnfocused) return;

    if (required && !value) {
      setError(name, { type: "required", message: "Required" });
      return;
    }

    if (type === "url") {
      try {
        const parsedURL = new URL(value);
        if (!["http:", "https:"].includes(parsedURL.protocol))
          throw new Error("Invalid protocol");
      } catch (error) {
        if (/^https?:\/\//.test(value as string)) {
          setError(name, {
            type: "url",
            message: "Invalid URL",
          });
          return;
        }
        setError(name, {
          type: "url",
          message: "Invalid URL - http(s):// prefix required",
        });
        return;
      }
    }
    // GitHub - OWNER/REPO
    else {
      if (!/^[\w-]+\/[\w-]+$/g.test(value)) {
        setError(name, {
          type: "github",
          message: "Must be in the format 'OWNER/REPO'",
        });
        return;
      }
    }

    clearErrors(name);
  }, [value, isUnfocused]);

  return (
    <Col xs={col_xs} sm={col_sm} className={`${padding} pt-1 pb-1 col-form`}>
      <FormGroup>
        <FormLabel
          text={type === "github" ? "Repository" : "URL"}
          tooltip={
            type === "github" ? (
              <>
                {"https://github.com/"}
                <span className="bold-underline">OWNER/REPO</span>
              </>
            ) : undefined
          }
          required={required !== false}
        />
        <InputGroup className="me-3">
          <FormControl
            defaultValue={value}
            onFocus={handleFocus}
            onBlur={handleBlur}
            onChange={(e) => setValue(name, e.target.value)}
            isInvalid={!!error}
          />
          {isUnfocused && value && !error && (
            <a href={link(type)} target="_blank">
              <Button variant="secondary" className="curved-right-only">
                <FontAwesomeIcon icon={faLink} />
              </Button>
            </a>
          )}
        </InputGroup>
        {error && (
          <small className="error-msg">{error["message"] || "err"}</small>
        )}
      </FormGroup>
    </Col>
  );
};

export default VersionWithLink;
