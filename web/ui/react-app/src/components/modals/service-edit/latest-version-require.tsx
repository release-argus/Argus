import { Accordion, FormGroup, Row } from "react-bootstrap";
import {
  DefaultDockerFilterRegistryType,
  DefaultDockerFilterType,
  DefaultLatestVersionFiltersType,
  DockerFilterRegistryType,
  DockerFilterType,
} from "types/config";
import { FC, memo, useEffect, useMemo } from "react";
import { FormItem, FormLabel, FormSelect } from "components/generic/form";
import { useFormContext, useWatch } from "react-hook-form";

import Command from "./command";

const DockerRegistryOptions = [
  { label: "Docker Hub", value: "hub" },
  { label: "GHCR", value: "ghcr" },
  { label: "Quay", value: "quay" },
];

type Props = {
  defaults?: DefaultLatestVersionFiltersType;
  hard_defaults?: DefaultLatestVersionFiltersType;
};

const EditServiceLatestVersionRequire: FC<Props> = ({
  defaults,
  hard_defaults,
}) => {
  const { setValue } = useFormContext();

  const defaultDockerRegistry: DefaultDockerFilterType["type"] =
    defaults?.docker?.type ?? hard_defaults?.docker?.type;
  const dockerRegistryOptions = useMemo(() => {
    if (defaultDockerRegistry === undefined) return DockerRegistryOptions;

    const defaultDockerRegistryLabel = DockerRegistryOptions.find(
      (option) =>
        option.value.toLowerCase() === defaultDockerRegistry.toLowerCase()
    );

    if (defaultDockerRegistryLabel)
      return [
        {
          value: "",
          label: `${defaultDockerRegistryLabel.label} (default)`,
        },
        ...DockerRegistryOptions,
      ];

    // Unknown default registry, return without this default.
    return DockerRegistryOptions;
  }, [defaultDockerRegistry]);
  const dockerRegistry: DockerFilterType["type"] = useWatch({
    name: "latest_version.require.docker.type",
  });
  const selectedDockerRegistry = (dockerRegistry ||
    defaultDockerRegistry) as DockerFilterRegistryType;
  const showUsernameField = (dockerRegistry || defaultDockerRegistry) === "hub";

  useEffect(() => {
    // Default to Docker Hub if no registry is selected and no default registry.
    if ((selectedDockerRegistry ?? "") === "")
      setValue("latest_version.require.docker.type", "hub");
  }, []);

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

          <FormGroup className="pt-1">
            <FormLabel
              text="Command"
              tooltip="Command to run before a release is considered usable and notified/shown in the UI"
            />
            <Command name="latest_version.require.command" />
          </FormGroup>

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
          {showUsernameField && (
            <FormItem
              key="username"
              name="latest_version.require.docker.username"
              col_sm={4}
              label="Username"
              defaultVal={
                (
                  defaults?.docker?.[
                    selectedDockerRegistry
                  ] as DefaultDockerFilterRegistryType
                )?.username ??
                (
                  hard_defaults?.docker?.[
                    selectedDockerRegistry
                  ] as DefaultDockerFilterRegistryType
                )?.username
              }
            />
          )}
          <FormItem
            name="latest_version.require.docker.token"
            key="token"
            col_sm={showUsernameField ? 8 : 12}
            label="Token"
            onRight={showUsernameField}
            defaultVal={
              (
                defaults?.docker?.[
                  selectedDockerRegistry
                ] as DefaultDockerFilterRegistryType
              )?.token ??
              (
                hard_defaults?.docker?.[
                  selectedDockerRegistry
                ] as DefaultDockerFilterRegistryType
              )?.token
            }
          />
        </Row>
      </Accordion.Body>
    </Accordion>
  );
};

export default memo(EditServiceLatestVersionRequire);
