import { Alert, Button } from "react-bootstrap";
import {
  DeployedVersionLookupEditType,
  LatestVersionLookupEditType,
  ServiceRefreshType,
} from "types/service-edit";
import { FC, useMemo, useState } from "react";
import {
  beautifyGoErrors,
  convertToQueryParams,
  fetchJSON,
  removeEmptyValues,
} from "utils";
import { faSpinner, faSync } from "@fortawesome/free-solid-svg-icons";
import { useFormContext, useWatch } from "react-hook-form";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { useErrors } from "hooks/errors";
import { useQuery } from "@tanstack/react-query";
import useValuesRefetch from "hooks/values-refetch";
import { useWebSocket } from "contexts/websocket";

interface Props {
  vType: 0 | 1; // 0: Latest, 1: Deployed
  serviceName: string;
  original?: LatestVersionLookupEditType | DeployedVersionLookupEditType;
}

/**
 * Returns the version with a button to refresh
 *
 * @param vType - 0: Latest, 1: Deployed
 * @param serviceName - The name of the service
 * @param original - The original values in the form
 * @returns The version with a button to refresh the version
 */
const VersionWithRefresh: FC<Props> = ({ vType, serviceName, original }) => {
  const [lastFetched, setLastFetched] = useState(0);
  const { monitorData } = useWebSocket();
  const { trigger } = useFormContext();
  const dataTarget = useMemo(
    () => (vType === 0 ? "latest_version" : "deployed_version"),
    []
  );
  const url: string | undefined = useWatch({ name: `${dataTarget}.url` });
  const dataTargetErrors = useErrors(dataTarget, true);
  const { data, refetchData } = useValuesRefetch(dataTarget);
  const { data: semanticVersioning, refetchData: refetchSemanticVersioning } =
    useValuesRefetch("options.semantic_versioning");

  const fetchVersionJSON = () =>
    fetchJSON<ServiceRefreshType>({
      url: `api/v1/${vType === 0 ? "latest" : "deployed"}_version/refresh${
        serviceName ? `/${encodeURIComponent(serviceName)}` : ""
      }?${
        data &&
        convertToQueryParams({
          params: { ...data, semantic_versioning: semanticVersioning },
          defaults: original,
        })
      }`,
    });

  const {
    data: versionData,
    isFetching,
    refetch: refetchVersion,
  } = useQuery({
    queryKey: [
      "version/refresh",
      dataTarget,
      { id: serviceName },
      {
        params: removeEmptyValues(data),
        semantic_versioning: semanticVersioning,
        original_data: removeEmptyValues(original ?? []),
      },
    ],
    queryFn: () => fetchVersionJSON(),
    enabled: false,
    initialData: {
      version: monitorData.service[serviceName]
        ? monitorData.service[serviceName]?.status?.[dataTarget]
        : "",
      error: "",
      timestamp: "",
    },
    notifyOnChangeProps: "all",
    staleTime: 0,
  });

  const refetch = async () => {
    // Prevent refetching too often
    const currentTime = Date.now();
    if (currentTime - lastFetched < 1000) return;

    // Ensure valid form
    const result = await trigger(dataTarget);
    if (!result) return;

    refetchSemanticVersioning();
    refetchData();
    // setTimeout to allow time for refetch setState's ^
    const timeout = setTimeout(() => {
      if (url) {
        refetchVersion();
        setLastFetched(currentTime);
      }
    });
    return () => clearTimeout(timeout);
  };

  const LoadingSpinner = (
    <FontAwesomeIcon icon={faSpinner} spin style={{ marginLeft: "0.5rem" }} />
  );

  return (
    <span style={{ alignItems: "center" }}>
      <span className="pt-1 pb-2" style={{ display: "flex" }}>
        {vType === 0 ? "Latest" : "Deployed"} version: {versionData.version}
        {data?.url !== "" && isFetching && LoadingSpinner}
        <Button
          variant="secondary"
          style={{ marginLeft: "auto", padding: "0 1rem" }}
          onClick={refetch}
          disabled={isFetching || !url}
        >
          <FontAwesomeIcon icon={faSync} style={{ paddingRight: "0.25rem" }} />
          Refresh
        </Button>
      </span>
      {(versionData.error || versionData.message) && (
        <span
          className="mb-2"
          style={{ width: "100%", wordBreak: "break-all" }}
        >
          <Alert variant="danger">
            Failed to refresh:
            <br />
            {beautifyGoErrors(
              (versionData.error || versionData.message) as string
            )}
          </Alert>
        </span>
      )}
      {dataTargetErrors && (
        <Alert
          variant="danger"
          style={{ paddingLeft: "2rem", marginBottom: "unset" }}
        >
          {Object.entries(dataTargetErrors).map(([key, error]) => (
            <li key={key}>
              {key}: {error}
            </li>
          ))}
        </Alert>
      )}
    </span>
  );
};

export default VersionWithRefresh;
