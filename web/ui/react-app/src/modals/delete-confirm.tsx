import { Button, Modal } from "react-bootstrap";
import { FC, useState } from "react";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faSpinner } from "@fortawesome/free-solid-svg-icons";

interface Props {
  onDelete: () => void;
  disabled?: boolean;
}

/**
 * Returns a delete confirmation modal
 *
 * @param onDelete - The function to call when the delete button is clicked
 * @param disabled - Whether the delete confirmation button is disabled
 * @returns A delete confirmation modal
 */
export const DeleteModal: FC<Props> = ({ onDelete, disabled }) => {
  const [modalShow, setModalShow] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const handleConfirm = async () => {
    setDeleting(true);
    // Call the onConfirm function
    onDelete();

    // Close the modal
    setModalShow(false);
  };
  return (
    <>
      <Button
        variant="danger"
        onClick={() => setModalShow(true)}
        disabled={disabled}
      >
        Delete
      </Button>
      <Modal
        show={modalShow}
        onHide={() => setModalShow(false)}
        style={{ backdropFilter: "blur(3px)" }}
        centered
      >
        <Modal.Header closeButton>
          <Modal.Title>Confirm Delete</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          Are you sure you want to delete this item? This action cannot be
          undone.
          {deleting && (
            <FontAwesomeIcon
              icon={faSpinner}
              spin
              style={{ marginLeft: "0.5rem" }}
            />
          )}
        </Modal.Body>
        <Modal.Footer>
          <Button variant="danger" onClick={handleConfirm} disabled={deleting}>
            Delete
          </Button>
          <Button
            variant="secondary"
            onClick={() => setModalShow(false)}
            disabled={deleting}
          >
            Cancel
          </Button>
        </Modal.Footer>
      </Modal>
    </>
  );
};
