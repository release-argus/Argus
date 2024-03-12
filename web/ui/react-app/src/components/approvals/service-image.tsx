import { FC, useMemo } from "react";
import {
  faCircleNotch,
  faWindowMaximize,
} from "@fortawesome/free-solid-svg-icons";

import { Card } from "react-bootstrap";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { ServiceSummaryType } from "types/summary";
import { faGithub } from "@fortawesome/free-brands-svg-icons";
import { useDelayedRender } from "hooks/delayed-render";

interface Props {
  service: ServiceSummaryType;
  visible: boolean;
}

export const ServiceImage: FC<Props> = ({ service, visible }) => {
  const delayedRender = useDelayedRender(500);
  const icon = useMemo(
    () => (service.type === "github" ? faGithub : faWindowMaximize),
    [service.type]
  );

  const imageStyles = {
    minWidth: "fit-content",
    height: "6rem",
  };

  const ServiceIcon = () => {
    if (service?.icon)
      return (
        <Card.Img
          variant="top"
          src={service.icon}
          alt={`${service.id} Image`}
          className="service-image"
        />
      );
    if (service?.loading)
      return (
        <div
          className="service-image"
          style={{ display: visible ? "inline" : "none" }}
        >
          {delayedRender(() => (
            <FontAwesomeIcon
              icon={faCircleNotch}
              style={{ ...imageStyles, padding: "0" }}
              className="service-image fa-spin"
            />
          ))}
        </div>
      );

    return (
      <FontAwesomeIcon
        icon={icon}
        style={imageStyles}
        className="service-image"
      />
    );
  };

  return (
    <div
      className="empty"
      style={{ height: "7rem", display: visible ? "flex" : "none" }}
    >
      <a
        href={service.icon_link_to ?? undefined}
        target="_blank"
        rel="noreferrer noopener"
        style={{ color: "inherit", display: "contents" }}
      >
        <ServiceIcon />
      </a>
    </div>
  );
};
