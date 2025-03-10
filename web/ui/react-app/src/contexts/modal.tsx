import { ModalType, ServiceModal, ServiceSummaryType } from 'types/summary';
import { ReactElement, ReactNode, createContext, useMemo } from 'react';

import ApprovalModal from 'modals/action-release';
import ServiceEditModal from 'modals/service-edit';
import useModal from 'hooks/modal';

interface ModalCtx {
	handleModal: (modalType: ModalType, service: ServiceSummaryType) => void;
	modal: ServiceModal;
}

/**
 * Provides modals to the application.
 *
 * @param modalType - The type of modal to display.
 * @param service - The service to display in the modal.
 * @returns The modal context.
 */
const ModalContext = createContext<ModalCtx>({
	handleModal: (_modalType: ModalType, _service: ServiceSummaryType) => {},
	modal: { actionType: '', service: { id: '', loading: true } },
});

interface Props {
	children: ReactNode;
}

/**
 * Provides modals to the application.
 *
 * @param props - The children to render.
 * @returns A Provider of modals to the application.
 */
const ModalProvider = (props: Props): ReactElement => {
	const { modal, handleModal } = useModal();
	const contextValue = useMemo(
		() => ({ handleModal, modal }),
		[handleModal, modal],
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
