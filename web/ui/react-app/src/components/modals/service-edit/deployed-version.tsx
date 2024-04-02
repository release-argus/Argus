import { Accordion, FormGroup, Row } from "react-bootstrap";
import { FC, memo, useEffect } from "react";
import {
  FormCheck,
  FormItem,
  FormKeyValMap,
  FormLabel,
} from "components/generic/form";
import { useFormContext, useWatch } from "react-hook-form";

import { BooleanWithDefault } from "components/generic";
import { DeployedVersionLookupType } from "types/config";
import VersionWithRefresh from "./version-with-refresh";

interface Props {
  serviceName: string;
  original?: DeployedVersionLookupType;
  defaults?: DeployedVersionLookupType;
  hard_defaults?: DeployedVersionLookupType;
}

/**
 * Returns the `deployed_version` form fields
 *
 * @param serviceName - The name of the service
 * @param original - The original values of the form
 * @param defaults - The default values
 * @param hard_defaults - The hard default
 * @returns The form fields for the `deployed_version`
 */
const EditServiceDeployedVersion: FC<Props> = ({
  serviceName,
  original,
  defaults,
  hard_defaults,
}) => {
  const { setValue } = useFormContext();

  // RegEx Template toggle
  const templateToggle: boolean = useWatch({
    name: "deployed_version.template_toggle",
  });
  useEffect(() => {
    // Clear the template if the toggle is false
    if (templateToggle === false) {
      setValue("deployed_version.regex_template", "");
      setValue("deployed_version.template_toggle", false);
    }
  }, [templateToggle]);

  return (
    <Accordion>
      <Accordion.Header>Deployed Version:</Accordion.Header>
      <Accordion.Body>
        <FormItem
          key="url"
          name="deployed_version.url"
          col_sm={12}
          label="URL"
          tooltip="URL to query for the version that's running"
          isURL
        />
        <BooleanWithDefault
          name="deployed_version.allow_invalid_certs"
          label="Allow Invalid Certs"
          defaultValue={
            defaults?.allow_invalid_certs || hard_defaults?.allow_invalid_certs
          }
        />
        <FormGroup className="pt-1 mb-2">
          <FormLabel text="Basic auth credentials" />
          <Row>
            <FormItem
              key="username"
              name="deployed_version.basic_auth.username"
              col_xs={6}
              label="Username"
              defaultVal={
                defaults?.basic_auth?.username ||
                hard_defaults?.basic_auth?.username
              }
            />
            <FormItem
              key="password"
              name="deployed_version.basic_auth.password"
              col_xs={6}
              label="Password"
              defaultVal={
                defaults?.basic_auth?.password ||
                hard_defaults?.basic_auth?.password
              }
              position="right"
            />
          </Row>
        </FormGroup>
        <FormKeyValMap name="deployed_version.headers" />
        <Row>
          <FormItem
            name="deployed_version.json"
            col_xs={6}
            label="JSON"
            tooltip={
              <>
                If the URL gives JSON, take the var at this location. e.g.
                <span className="bold-underline">data.version</span>
              </>
            }
            defaultVal={defaults?.json || hard_defaults?.json}
          />
          <FormItem
            name="deployed_version.regex"
            required={templateToggle ? "Required for template" : undefined}
            col_xs={4}
            col_sm={5}
            label="RegEx"
            tooltip={
              <>
                RegEx to extract the version from the URL, e.g.
                <span className="bold-underline">v([0-9.]+)</span>
              </>
            }
            defaultVal={defaults?.regex || hard_defaults?.regex}
            isRegex
            position="middle"
          />
          <FormCheck
            name={`deployed_version.template_toggle`}
            size="lg"
            label="T"
            tooltip="Use the RegEx to create a template ($1,$2,etc to reference capture groups)"
            smallLabel
            col_sm={1}
            col_xs={2}
            position="right"
          />
          {templateToggle && (
            <FormItem
              name="deployed_version.regex_template"
              col_sm={12}
              label="RegEx Template"
              tooltip="Use the RegEx to create a template ($1,$2,etc to reference capture groups)"
            />
          )}
        </Row>
        <VersionWithRefresh
          vType={1}
          serviceName={serviceName}
          original={original}
        />
      </Accordion.Body>
    </Accordion>
  );
};

export default memo(EditServiceDeployedVersion);
