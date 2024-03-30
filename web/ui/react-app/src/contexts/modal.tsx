import { ModalType, ServiceModal, ServiceSummaryType } from "types/summary";
import { ReactElement, ReactNode, createContext, useMemo } from "react";

import ApprovalModal from "modals/action-release";
import ServiceEditModal from "modals/service-edit";
import useModal from "hooks/modal";

interface ModalCtx {
  handleModal: (modalType: ModalType, service: ServiceSummaryType) => void;
  modal: ServiceModal;
}

/**
 * The modal context, which provides modals to the application.
 * 
 * @param modalType - The type of modal to display
 * @param service - The service to display in the modal
 * @returns The modal context
 */
const ModalContext = createContext<ModalCtx>({
  // eslint-disable-next-line @typescript-eslint/no-empty-function, @typescript-eslint/no-unused-vars
  handleModal: (modalType: ModalType, service: ServiceSummaryType) => {},
  modal: { actionType: "", service: { id: "", loading: true } },
});

interface Props {
  children: ReactNode;
}

/**
 * Returns the modal provider, which provides modals to the application.
 *
 * @param props - The children to render
 * @returns The modal provider, which provides modals to the application.
 */
const ModalProvider = (props: Props): ReactElement => {
  const { modal, handleModal } = useModal();
  const contextValue = useMemo(
    () => ({ handleModal, modal }),
    [handleModal, modal]
  );

  return (
    <ModalContext.Provider value={contextValue}>
      <ApprovalModal />
      <ServiceEditModal />
      {props.children}
    </ModalContext.Provider>
  );
};

export { ModalContext, ModalProvider };
