import { Button, Card } from "react-bootstrap";
import { FC, memo, useCallback, useContext, useMemo, useState } from "react";
import { ModalType, ServiceSummaryType } from "types/summary";
import { ServiceImage, ServiceInfo, UpdateInfo } from "components/approvals";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { ModalContext } from "contexts/modal";
import { faPen } from "@fortawesome/free-solid-svg-icons";

interface Props {
  service: ServiceSummaryType;
  editable: boolean;
}

const Service: FC<Props> = ({ service, editable = false }) => {
  const [showUpdateInfo, setShowUpdateInfoMain] = useState(false);

  const setShowUpdateInfo = useCallback(() => {
    setShowUpdateInfoMain((prevState) => !prevState);
  }, []);
  const { handleModal } = useContext(ModalContext);

  const showModal = useMemo(
    () => (type: ModalType, service: ServiceSummaryType) => {
      handleModal(type, service);
    },
    []
  );

  const updateAvailable = useMemo(
    (): boolean =>
      (service?.status?.deployed_version ?? undefined) !==
      (service?.status?.latest_version ?? undefined),
    [service?.status?.latest_version, service?.status?.deployed_version]
  );

  const updateSkipped = useMemo(
    (): boolean =>
      updateAvailable &&
      service?.status?.approved_version ===
        `SKIP_${service?.status?.latest_version}`,
    [
      updateAvailable,
      service?.status?.approved_version,
      service?.status?.latest_version,
    ]
  );

  return (
    <Card key={service.id} bg="secondary" className={"service shadow"}>
      <Card.Title className="service-title" key={service.id + "-title"}>
        <a
          className="same-color"
          href={service.url}
          target="_blank"
          rel="noreferrer noopener"
          style={{ height: "100% !important" }}
        >
          <strong>{service.id}</strong>
        </a>
        {editable && (
          <Button
            className="btn-icon-center"
            size="sm"
            variant="secondary"
            onClick={() => showModal("EDIT", service)}
            style={{
              height: "1.5rem",
              width: "1.5rem",

              // lay it on top
              zIndex: 1,
              position: "absolute",
              top: "0.5rem",
              right: "0.5rem",
            }}
          >
            <FontAwesomeIcon icon={faPen} className="fa-sm" />
          </Button>
        )}
      </Card.Title>

      <Card
        key={service.id}
        bg="secondary"
        className={`service-inner ${
          service.active === false ? "service-disabled" : ""
        }`}
      >
        <UpdateInfo
          service={service}
          visible={updateAvailable && showUpdateInfo && !updateSkipped}
        />
        <ServiceImage
          service={service}
          visible={!(updateAvailable && showUpdateInfo && !updateSkipped)}
        />
        <ServiceInfo
          service={service}
          setShowUpdateInfo={setShowUpdateInfo}
          updateAvailable={updateAvailable}
          updateSkipped={updateSkipped}
        />
      </Card>
    </Card>
  );
};

export default memo(Service);
