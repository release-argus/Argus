import { Alert, Button } from "react-bootstrap";
import { FC, useMemo } from "react";
import {
  LatestVersionLookupEditType,
  ServiceRefreshType,
} from "types/service-edit";
import { convertToQueryParams, fetchJSON, stringifyQueryParam } from "utils";
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

const VersionWithRefresh: FC<Props> = ({ vType, serviceName, original }) => {
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
          params: data,
          defaults: original,
        })
      }${stringifyQueryParam("semantic_versioning", semanticVersioning, true)}`
    );

  const {
    data: versionData,
    isFetching,
    refetch: refetchVersion,
  } = useQuery(
    [
      "refresh",
      dataTarget,
      { id: serviceName },
      {
        params: data,
        semantic_versioning: semanticVersioning,
        original_data: original,
      },
    ],
    () => fetchVersionJSON(),
    {
      initialData: {
        version: monitorData.service[serviceName]
          ? monitorData.service[serviceName]?.status?.[dataTarget]
          : "",
        error: "",
        timestamp: "",
      },
      enabled: !invalidURL && !!data?.url,
      notifyOnChangeProps: "all",
    }
  );

  const refetch = async () => {
    if (!invalidURL && !!data?.url) {
      refetchSemanticVersioning();
      refetchData();
      // setTimeoout to allow time for refetches ^
      setTimeout(() => {
        refetchVersion();
      });
    }
  };

  const LoadingSpinner = useMemo(
    () => (
      <FontAwesomeIcon icon={faSpinner} spin style={{ marginLeft: "0.5rem" }} />
    ),
    []
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
            {versionData.error.replaceAll("\\", "\n")}
          </Alert>
        </span>
      )}
    </span>
  );
};

export default VersionWithRefresh;
