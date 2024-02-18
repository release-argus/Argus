import {
  Button,
  Card,
  Container,
  ListGroup,
  OverlayTrigger,
  Tooltip,
} from "react-bootstrap";
import { FC, useCallback, useContext, useMemo } from "react";
import { ModalType, ServiceSummaryType } from "types/summary";
import {
  faArrowRotateRight,
  faCheck,
  faInfo,
  faInfoCircle,
  faSatelliteDish,
  faTimes,
} from "@fortawesome/free-solid-svg-icons";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { ModalContext } from "contexts/modal";
import { formatRelative } from "date-fns";

interface Props {
  service: ServiceSummaryType;
  updateAvailable: boolean;
  updateSkipped: boolean;
  setShowUpdateInfo: () => void;
}

export const ServiceInfo: FC<Props> = ({
  service,
  updateAvailable,
  updateSkipped,
  setShowUpdateInfo,
}) => {
  const { handleModal } = useContext(ModalContext);

  const showModal = useCallback(
    (type: ModalType, service: ServiceSummaryType) => {
      handleModal(type, service);
    },
    [handleModal]
  );

  // If version hasn't been found or a new version has been found
  const serviceWarning = useMemo(
    () =>
      service?.status?.deployed_version === undefined ||
      service?.status?.deployed_version === "" ||
      (updateAvailable && !updateSkipped),
    [service, updateAvailable, updateSkipped]
  );

  const updateApproved = useMemo(
    () =>
      service?.status?.latest_version !== undefined &&
      service.status.latest_version === service?.status?.approved_version,
    [service]
  );

  const deployedVersionIcon = service.has_deployed_version ? (
    <OverlayTrigger
      key="deployed-service"
      placement="top"
      delay={{ show: 500, hide: 500 }}
      overlay={
        <Tooltip id={`tooltip-deployed-service`}>
          of the deployed {service.id}
        </Tooltip>
      }
    >
      <FontAwesomeIcon
        className="same-color"
        style={{ paddingLeft: "0.5rem", paddingBottom: "0.1rem" }}
        icon={faSatelliteDish}
      />
    </OverlayTrigger>
  ) : null;

  const skippedVersionIcon =
    updateSkipped && service.status?.approved_version ? (
      <OverlayTrigger
        key="skipped-version"
        placement="top"
        delay={{ show: 500, hide: 500 }}
        overlay={
          <Tooltip id={`tooltip-skipped-version`}>
            Skipped {service.status.approved_version.slice("SKIP_".length)}
          </Tooltip>
        }
      >
        <FontAwesomeIcon
          icon={faInfoCircle}
          style={{ paddingLeft: "0.5rem", paddingBottom: "0.1rem" }}
        />
      </OverlayTrigger>
    ) : null;

  const actionReleaseButton =
    (service.webhook || service.command) &&
    (!updateAvailable || updateSkipped) ? (
      <OverlayTrigger
        key="resend"
        placement="top"
        delay={{ show: 500, hide: 500 }}
        overlay={
          <Tooltip id={`tooltip-resend`}>
            {updateSkipped
              ? "Approve this release"
              : `Resend the ${service.webhook ? "WebHooks" : "Commands"}`}
          </Tooltip>
        }
      >
        <Button
          variant="secondary"
          size="sm"
          onClick={() => showModal(updateSkipped ? "SEND" : "RESEND", service)}
          disabled={service.loading || service.active === false}
        >
          <FontAwesomeIcon
            icon={updateSkipped ? faCheck : faArrowRotateRight}
          />
        </Button>
      </OverlayTrigger>
    ) : null;

  return (
    <Container
      style={{
        padding: "0px",
      }}
      className={serviceWarning ? "service-warning rounded-bottom" : "default"}
    >
      <ListGroup className="list-group-flush">
        {updateAvailable && !updateSkipped ? (
          <>
            <ListGroup.Item
              key="update-available"
              className={"service-item update-options service-warning"}
              variant="secondary"
            >
              {updateApproved && (service.webhook || service.command)
                ? `${service.webhook ? "WebHooks" : "Commands"} already sent:`
                : "Update available:"}
            </ListGroup.Item>
            <ListGroup.Item
              key="update-buttons"
              className={"service-item update-options service-warning"}
              variant="secondary"
              style={{ paddingTop: "0.25rem" }}
            >
              <Button
                key="details"
                className="btn-flex btn-update-action"
                variant="primary"
                onClick={() => setShowUpdateInfo()}
              >
                <FontAwesomeIcon icon={faInfo} />
              </Button>
              <Button
                key="approve"
                className={`btn-flex btn-update-action${
                  service.webhook || service.command ? "" : " hiddenElement"
                }`}
                variant="success"
                onClick={() =>
                  showModal(updateApproved ? "RESEND" : "SEND", service)
                }
                disabled={!(service.webhook || service.command)}
              >
                <FontAwesomeIcon icon={faCheck} />
              </Button>
              <Button
                key="reject"
                className="btn-flex btn-update-action"
                variant="danger"
                onClick={() =>
                  showModal(
                    service.webhook || service.command ? "SKIP" : "SKIP_NO_WH",
                    service
                  )
                }
              >
                <FontAwesomeIcon icon={faTimes} color="white" />
              </Button>
            </ListGroup.Item>
          </>
        ) : (
          <ListGroup.Item
            key="deployed_v"
            variant={serviceWarning ? "warning" : "secondary"}
            className={
              "service-item" +
              (service.webhook || service.command ? "" : " justify-left")
            }
          >
            <div style={{ margin: 0 }}>
              <>
                Current version:
                {deployedVersionIcon}
                {skippedVersionIcon}
              </>
              <br />
              <div style={{ display: "flex", margin: 0 }}>
                <OverlayTrigger
                  key="deployed-version"
                  placement="top"
                  delay={{ show: 500, hide: 500 }}
                  overlay={
                    service?.status?.deployed_version_timestamp ? (
                      <Tooltip id={`tooltip-deployed-version`}>
                        <>
                          {formatRelative(
                            new Date(service.status.deployed_version_timestamp),
                            new Date()
                          )}
                        </>
                      </Tooltip>
                    ) : (
                      <>Unknown</>
                    )
                  }
                >
                  <p style={{ margin: 0 }}>
                    {service?.status?.deployed_version
                      ? service.status.deployed_version
                      : "Unknown"}{" "}
                  </p>
                </OverlayTrigger>
              </div>
            </div>
            {actionReleaseButton}
          </ListGroup.Item>
        )}
      </ListGroup>
      <Card.Footer
        className={
          serviceWarning || !service?.status?.last_queried
            ? "service-warning rounded-bottom"
            : ""
        }
      >
        <small
          className={
            "text-muted same-color" +
            (serviceWarning ? " service-warning rounded-bottom" : "")
          }
        >
          {service?.status?.last_queried ? (
            <>
              queried{" "}
              {formatRelative(
                new Date(service.status.last_queried),
                new Date()
              )}
            </>
          ) : service.loading ? (
            "loading"
          ) : (
            "no successful queries"
          )}
        </small>
      </Card.Footer>
    </Container>
  );
};
