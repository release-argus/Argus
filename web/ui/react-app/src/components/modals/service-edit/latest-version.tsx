import { Accordion, Row } from "react-bootstrap";
import {
  DefaultLatestVersionLookupType,
  LatestVersionLookupType,
} from "types/config";
import { FC, memo, useMemo } from "react";
import { FormItem, FormSelect } from "components/generic/form";

import { BooleanWithDefault } from "components/generic";
import EditServiceLatestVersionRequire from "./latest-version-require";
import FormURLCommands from "./latest-version-urlcommands";
import { LatestVersionLookupEditType } from "types/service-edit";
import VersionWithRefresh from "./version-with-refresh";
import { firstNonDefault } from "components/modals/service-edit/util";
import { useWatch } from "react-hook-form";

interface Props {
  serviceName: string;
  original?: LatestVersionLookupEditType;
  defaults?: DefaultLatestVersionLookupType;
  hard_defaults?: DefaultLatestVersionLookupType;
}

/**
 * Returns the `latest_version` form fields
 *
 * @param serviceName - The name of the service
 * @param original - The original values in the form
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for the `latest_version`
 */
const EditServiceLatestVersion: FC<Props> = ({
  serviceName,
  original,
  defaults,
  hard_defaults,
}) => {
  const latestVersionTypeOptions: {
    label: string;
    value: NonNullable<LatestVersionLookupType["type"]>;
  }[] = [
    { label: "GitHub", value: "github" },
    { label: "URL", value: "url" },
  ];

  const latestVersionType: LatestVersionLookupType["type"] = useWatch({
    name: `latest_version.type`,
  });

  const convertedDefaults = useMemo(
    () => ({
      access_token: firstNonDefault(
        defaults?.access_token,
        hard_defaults?.access_token
      ),
      allow_invalid_certs:
        defaults?.allow_invalid_certs ?? hard_defaults?.allow_invalid_certs,
      use_prerelease: defaults?.use_prerelease ?? hard_defaults?.use_prerelease,
    }),
    [defaults, hard_defaults]
  );

  return (
    <Accordion>
      <Accordion.Header>Latest Version:</Accordion.Header>
      <Accordion.Body>
        <Row>
          <FormSelect
            name="latest_version.type"
            col_xs={4}
            col_sm={4}
            label="Type"
            options={latestVersionTypeOptions}
          />
          <FormItem
            name="latest_version.url"
            required
            col_sm={8}
            col_xs={8}
            label={latestVersionType === "github" ? "Repo" : "URL"}
            position="right"
          />
          {latestVersionType === "github" ? (
            <>
              <FormItem
                key="access_token"
                name="latest_version.access_token"
                col_sm={12}
                label="Access Token"
                tooltip="GitHub Personal Access Token to handle possible rate limits and/or private repos"
                defaultVal={convertedDefaults.access_token}
                isURL={latestVersionType !== "github"}
              />
              <BooleanWithDefault
                name="latest_version.use_prerelease"
                label="Use pre-releases"
                tooltip="Include releases marked 'Pre-release' in the latest version check"
                defaultValue={convertedDefaults.use_prerelease}
              />
            </>
          ) : (
            <BooleanWithDefault
              name="latest_version.allow_invalid_certs"
              label="Allow Invalid Certs"
              defaultValue={convertedDefaults.allow_invalid_certs}
            />
          )}
          <FormURLCommands />
          <EditServiceLatestVersionRequire
            defaults={defaults?.require}
            hard_defaults={hard_defaults?.require}
          />

          <VersionWithRefresh
            vType={0}
            serviceName={serviceName}
            original={original}
          />
        </Row>
      </Accordion.Body>
    </Accordion>
  );
};

export default memo(EditServiceLatestVersion);
