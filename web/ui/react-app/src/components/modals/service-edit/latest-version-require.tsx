import { Accordion, Form, Row } from "react-bootstrap";
import { FormItem, FormLabel, FormSelect } from "components/generic/form";
import { useFormContext, useWatch } from "react-hook-form";

import Command from "./command";
import { memo } from "react";

const EditServiceLatestVersionRequire = () => {
  const { control } = useFormContext();
  const dockerRegistry = useWatch({
    control,
    name: "latest_version.require.docker.type",
  });
  const dockerRegistryOptions = [
    { label: "Docker Hub", value: "hub" },
    { label: "GHCR", value: "ghcr" },
    { label: "Quay", value: "quay" },
  ];

  return (
    <Accordion style={{ marginBottom: "0.5rem" }}>
      <Accordion.Header>Require:</Accordion.Header>
      <Accordion.Body>
        <Row>
          <FormItem
            name="latest_version.require.regex_content"
            col_xs={6}
            label={"RegEx Content"}
            tooltip="GitHub=assets must contain a match, URL=webpage must contain a match"
            isRegex
          />
          <FormItem
            name="latest_version.require.regex_version"
            col_xs={6}
            label={"RegEx Version"}
            tooltip="Version found must match, e.g. exclude '*-beta' versions with '^[0-9.]+$'"
            isRegex
            onRight
          />

          <Form.Group className="pt-1">
            <FormLabel
              text="Command"
              tooltip="Command to run before a release is considered usable and notified/shown in the UI"
            />
            <Command name="latest_version.require.command" />
          </Form.Group>

          <hr />
          <FormLabel text="Docker" />
          <FormSelect
            name="latest_version.require.docker.type"
            col_xs={12}
            col_sm={12}
            label="Type"
            options={dockerRegistryOptions}
          />
          <FormItem
            name="latest_version.require.docker.image"
            label="Image"
            col_xs={6}
            onRight={false}
          />
          <FormItem
            name="latest_version.require.docker.tag"
            col_xs={6}
            label="Tag"
            onRight
          />
          {dockerRegistry === "hub" && (
            <FormItem
              key="username"
              name="latest_version.require.docker.username"
              col_sm={4}
              label="Username"
            />
          )}
          <FormItem
            name="latest_version.require.docker.token"
            key="token"
            col_sm={dockerRegistry === "hub" ? 8 : 12}
            label="Token"
            onRight={dockerRegistry === "hub"}
          />
        </Row>
      </Accordion.Body>
    </Accordion>
  );
};

export default memo(EditServiceLatestVersionRequire);
