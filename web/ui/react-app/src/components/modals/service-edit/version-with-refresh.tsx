import { Alert, Button } from "react-bootstrap";
import { FC, useMemo, useState } from "react";
import {
  LatestVersionLookupEditType,
  ServiceRefreshType,
} from "types/service-edit";
import { convertToQueryParams, fetchJSON, removeEmptyValues } from "utils";
import { faSpinner, faSync } from "@fortawesome/free-solid-svg-icons";

import { DeployedVersionLookupType } from "types/config";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { useFormContext } from "react-hook-form";
import { useQuery } from "@tanstack/react-query";
import useValuesRefetch from "hooks/values-refetch";
import { useWebSocket } from "contexts/websocket";

interface Props {
  vType: 0 | 1; // 0: Latest, 1: Deployed
  serviceName: string;
  original?: LatestVersionLookupEditType | DeployedVersionLookupType;
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
  const { getFieldState, formState } = useFormContext();
  const dataTarget = useMemo(
    () => (vType === 0 ? "latest_version" : "deployed_version"),
    []
  );
  const { data, refetchData } = useValuesRefetch(dataTarget);
  const { data: semanticVersioning, refetchData: refetchSemanticVersioning } =
    useValuesRefetch("options.semantic_versioning");
  const { error: invalidURL } = getFieldState(dataTarget + ".url", formState);

  const fetchVersionJSON = () =>
    fetchJSON<ServiceRefreshType>(
      `api/v1/${vType === 0 ? "latest" : "deployed"}_version/refresh${
        serviceName ? `/${encodeURIComponent(serviceName)}` : ""
      }?${
        data &&
        convertToQueryParams({
          params: { ...data, semantic_versioning: semanticVersioning },
          defaults: original,
        })
      }`
    );

  const {
    data: versionData,
    isFetching,
    isStale,
    refetch: refetchVersion,
  } = useQuery({
    queryKey: [
      "version/refresh",
      dataTarget,
      { id: serviceName },
      {
        params: JSON.stringify(removeEmptyValues(data)),
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

    if (isStale && !invalidURL && !!data?.url) {
      refetchSemanticVersioning();
      refetchData();
      // setTimeout to allow time for refetches ^
      setTimeout(() => {
        refetchVersion();
      });
      setLastFetched(currentTime);
    }
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
          disabled={isFetching || invalidURL !== undefined || !data?.url}
        >
          <FontAwesomeIcon icon={faSync} style={{ paddingRight: "0.25rem" }} />
          Refresh
        </Button>
      </span>
      {versionData.error && (
        <span
          className="mb-2"
          style={{ width: "100%", wordBreak: "break-all" }}
        >
          <Alert variant="danger">
            Failed to refresh:
            <br />
            {
              versionData.error
                .replaceAll(/\\([ \t])/g, "\n$1") // \ + space/tab -> newline
                .replaceAll(`\\n`, "\n") // \n -> newline
                .replaceAll(`\\"`, `"`) // \" -> "
                .replaceAll(`\\\\`, `\\`) // \\ -> \
                .replaceAll(/\\$/g, "\n") // \ + end of string -> newline
            }
          </Alert>
        </span>
      )}
    </span>
  );
};

export default VersionWithRefresh;
