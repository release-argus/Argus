import { ModalType, ServiceSummaryType, WebHookModal } from "types/summary";
import { ReactElement, ReactNode, createContext } from "react";

import ApprovalModal from "modals/webhook";
import useModal from "hooks/modal";

interface ModalCtx {
  handleModal: (modalType: ModalType, service: ServiceSummaryType) => void;
  modal: WebHookModal;
}

const ModalContext = createContext<ModalCtx>({
  // eslint-disable-next-line @typescript-eslint/no-empty-function, @typescript-eslint/no-unused-vars
  handleModal: (modalType: ModalType, service: ServiceSummaryType) => {},
  modal: { type: "", service: { id: "", loading: true } },
});
const { Provider } = ModalContext;

interface Props {
  children: ReactNode;
}

const ModalProvider = (props: Props): ReactElement => {
  const { modal, handleModal } = useModal();

  return (
    <Provider
      value={{
        handleModal,
        modal,
      }}
    >
      <ApprovalModal />
      {props.children}
    </Provider>
  );
};

export { ModalContext, ModalProvider };
