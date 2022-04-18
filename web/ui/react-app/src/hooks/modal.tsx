import { ModalType, ServiceSummaryType, WebHookModal } from "types/summary";

import { useState } from "react";

const useModal = () => {
  const [modal, setModal] = useState<WebHookModal>({
    type: "",
    service: { id: "", loading: true },
  });

  const handleModal = (type: ModalType, service: ServiceSummaryType) => {
    setModal({ type, service });
  };

  return { modal, handleModal };
};

export default useModal;
