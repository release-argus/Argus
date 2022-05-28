import { Button, Card, Container, Placeholder } from "react-bootstrap";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { ModalType } from "types/summary";
import { faSquareFull } from "@fortawesome/free-solid-svg-icons";

interface params {
  modalType: ModalType;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  delayedRender: any;
}

export const Loading = ({ modalType, delayedRender }: params) => {
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
