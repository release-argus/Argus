import { Button, Card, Container, Placeholder } from "react-bootstrap";

import { FC } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { ModalType } from "types/summary";
import { faSquareFull } from "@fortawesome/free-solid-svg-icons";

interface Props {
  modalType: ModalType;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  delayedRender: any;
}

/**
 * Renders a loading placeholder of the action-release modal
 *
 * @param modalType - The type of the modal
 * @param delayedRender - The delayed render function
 * @returns A loading component that has the confirmation button disabled if the modal type is "SKIP"
 */
export const Loading: FC<Props> = ({ modalType, delayedRender }) => {
  return (
    <Card bg="secondary" className={"no-margin service"}>
      <Card.Title className="title">
        <Container fluid style={{ paddingLeft: "0px" }}>
          {delayedRender(() => (
            <Placeholder xs={4} />
          ))}
        </Container>

        {modalType !== "SKIP" && (
          <Button
            variant="secondary"
            size="sm"
            className="float-end"
            // Disable if success or waiting send response
            disabled
          >
            <FontAwesomeIcon icon={faSquareFull} />
          </Button>
        )}
      </Card.Title>
    </Card>
  );
};
