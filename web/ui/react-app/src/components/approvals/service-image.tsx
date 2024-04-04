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

/**
 * Returns the service's image, with a loading spinner if the image is not loaded yet
 * and a link to the service. If the service has no icon, the service type icon (github/url) is displayed.
 *
 * @param service - The service the image belongs to
 * @param visible - Whether the image should be visible
 * @returns A component that displays the image of the service
 */
export const ServiceImage: FC<Props> = ({ service, visible }) => {
  const delayedRender = useDelayedRender(500);

  const imageStyles = {
    minWidth: "fit-content",
    height: "6rem",
  };

  const icon = useMemo(() => {
    // URL icon
    if (service?.icon)
      return (
        <Card.Img
          variant="top"
          src={service.icon}
          alt={`${service.id} Image`}
          className="service-image"
        />
      );

    // Loading spinner
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

    // Default icon
    return (
      <FontAwesomeIcon
        icon={service.type === "github" ? faGithub : faWindowMaximize}
        style={imageStyles}
        className="service-image"
      />
    );
  }, [service.icon]);

  return (
    <div
      className="empty"
      style={{ height: "7rem", display: visible ? "flex" : "none" }}
    >
      <a
        href={service.icon_link_to || undefined}
        target="_blank"
        rel="noreferrer noopener"
        style={{ color: "inherit", display: "contents" }}
      >
        {icon}
      </a>
    </div>
  );
};
