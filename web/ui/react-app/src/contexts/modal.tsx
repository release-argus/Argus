import {
	createContext,
	type Dispatch,
	type FC,
	type ReactNode,
	type SetStateAction,
} from 'react';
import useModal from '@/hooks/use-modal';
import { TooltipProviderGlobal } from '@/hooks/use-tooltip';
import ActionReleaseModal from '@/modals/action-release';
import ServiceEditModal from '@/modals/service-edit';
import type { ServiceModal } from '@/utils/api/types/config/summary';

type ModalContextProps = {
	/* The function to handle the modal. */
	setModal: Dispatch<SetStateAction<ServiceModal>>;
	/* The modal to display. */
	modal: ServiceModal;
};

/**
 * Provides modals to the app.
 *
 * @param modalType - The modal type to display.
 * @param service - The service to display in the modal.
 * @returns The modal context.
 */
const ModalContext = createContext<ModalContextProps>({
	modal: { actionType: '', service: { id: '', loading: true } },
	setModal: () => {
		/* noop */
	},
});

type ModalProviderProps = {
	/* The content to wrap. */
	children: ReactNode;
};

/**
 * @returns A Provider of modals to the app.
 */
const ModalProvider: FC<ModalProviderProps> = ({ children }) => {
	const contextValue = useModal();

	return (
		<ModalContext value={contextValue}>
			<TooltipProviderGlobal>
				<ActionReleaseModal />
				<ServiceEditModal />
				{children}
			</TooltipProviderGlobal>
		</ModalContext>
	);
};

export { ModalContext, ModalProvider };
