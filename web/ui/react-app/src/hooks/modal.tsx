import { ModalType, ServiceModal, ServiceSummaryType } from "types/summary";

import { useState } from "react";

/**
 * useModal is a hook to handle the state of the ServiceModa
 */
const useModal = () => {
  const [modal, setModal] = useState<ServiceModal>({
    actionType: "",
    service: { id: "", loading: true },
  });

  const handleModal = (actionType: ModalType, service: ServiceSummaryType) => {
    setModal({ actionType, service });
  };

  return { modal, handleModal };
};

export default useModal;
