import { Alert, Button } from "react-bootstrap";
import { FC, useMemo, useState } from "react";
import { convertToQueryParams, fetchJSON } from "utils";
import {
  faCheckCircle,
  faCircleXmark,
  faSpinner,
  faSync,
} from "@fortawesome/free-solid-svg-icons";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { NotifyType } from "types/config";
import { convertNotifyToAPI } from "components/modals/service-edit/util/ui-api-conversions";
import { useFormState } from "react-hook-form";
import { useQuery } from "@tanstack/react-query";
import useValuesRefetch from "hooks/values-refetch";

const Result: FC<{ status: "pending" | "success" | "error"; err?: string }> = ({
  status,
  err,
}) => {
  if (status === "pending") return <></>;
  return (
    <span className="mb-2" style={{ width: "100%", wordBreak: "break-all" }}>
      <Alert variant={err || status === "error" ? "danger" : "success"}>
        {err || status === "error"
          ? (err ?? "error") // Styling for verify errs
              .replaceAll(/\\([ \t])/g, "\n$1") // \ + space/tab -> newline
              .replaceAll(`\\n`, "\n") // \n -> newline
              .replaceAll(`\\"`, `"`) // \" -> "
              .replaceAll(`\\\\`, `\\`) // \\ -> \
              .replaceAll(/\\$/g, "\n") // \ + end of string -> newline
          : "Success!"}
      </Alert>
    </span>
  );
};

interface Props {
  path: string;
  original?: NotifyType;
  extras?: {
    service_name?: string;
    service_url?: string;
    web_url?: string;
  };
}

const TestNotify: FC<Props> = ({ path, original, extras }) => {
  const { isValid } = useFormState({ name: path });
  const [lastFetched, setLastFetched] = useState(0);
  const { data, refetchData } = useValuesRefetch(path, true);

  const fetchTestNotifyJSON = () =>
    fetchJSON<{ message: string }>(
      `api/v1/notify/test?${convertToQueryParams({
        params: {
          notify_name: original?.name,
          ...extras,

          options: {},
          url_fields: {},
          params: {},

          ...convertNotifyToAPI(data),
        },
        defaults: original,
      })}`
    );

  const {
    data: testData,
    isFetching,
    refetch: testRefetch,
    status: testStatus,
  } = useQuery({
    queryKey: [
      "test_notify",
      { id: original?.id },
      {
        params: convertNotifyToAPI(data),
      },
    ],
    queryFn: () => fetchTestNotifyJSON(),
    enabled: false,
    notifyOnChangeProps: "all",
    retry: false,
    staleTime: 0,
  });

  const refetch = async () => {
    // Prevent refetching too often
    const currentTime = Date.now();
    if (currentTime - lastFetched < 2000) return;

    if (isValid) {
      refetchData();
      // setTimeoout to allow for refetches ^
      setTimeout(() => {
        testRefetch();
      });
      setLastFetched(currentTime);
    }
  };

  const ResultIcon = useMemo(() => {
    if (isFetching)
      return (
        <FontAwesomeIcon
          icon={faSpinner}
          spin
          style={{ marginLeft: "0.5rem" }}
        />
      );
    if (testStatus !== "pending") {
      const err = testData?.message !== undefined || testStatus === "error";
      return (
        <FontAwesomeIcon
          icon={err ? faCircleXmark : faCheckCircle}
          className={`text-${err ? "danger" : "success"}`}
        />
      );
    }
    return <></>;
  }, [isFetching, testStatus, testData]);

  return (
    <span style={{ alignItems: "center" }}>
      <span className="pt-1 pb-2" style={{ display: "flex" }}>
        {ResultIcon}
        <Button
          variant="secondary"
          style={{ marginLeft: "auto", padding: "0 1rem" }}
          onClick={refetch}
          disabled={!isValid || isFetching}
        >
          <FontAwesomeIcon icon={faSync} style={{ paddingRight: "0.25rem" }} />
          Send Test Message
        </Button>
      </span>
      <Result status={testStatus} err={testData?.message} />
    </span>
  );
};

export default TestNotify;
