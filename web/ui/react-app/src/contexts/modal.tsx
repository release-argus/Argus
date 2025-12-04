import {
	createContext,
	type Dispatch,
	type FC,
	type ReactNode,
	type SetStateAction,
	useCallback,
	useMemo,
	useState,
} from 'react';
import { TooltipProviderGlobal } from '@/hooks/use-tooltip';
import ActionReleaseModal from '@/modals/action-release';
import ServiceEditModal from '@/modals/service-edit';
import type { ServiceModal } from '@/utils/api/types/config/summary';

type ModalProviderContextProps = {
	/* The function to handle the modal. */
	setModal: Dispatch<SetStateAction<ServiceModal>>;
	/* The modal to display. */
	modal: ServiceModal;
	/* The function to hide the modal. */
	hideModal: () => void;
};

/**
 * Provides modals to the app.
 *
 * @param modalType - The modal type to display.
 * @param service - The service to display in the modal.
 * @returns The modal context.
 */
const ModalProviderContext = createContext<ModalProviderContextProps>({
	hideModal: () => {
		/* noop */
	},
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
	const [modal, setModal] = useState<ServiceModal>({
		actionType: '',
		service: { id: '', loading: true },
	});

	const hideModal = useCallback(() => {
		setModal({ actionType: '', service: { id: '', loading: true } });
	}, []);

	const contextValue = useMemo(
		() => ({ hideModal, modal, setModal }),
		[hideModal, modal],
	);

	return (
		<ModalProviderContext value={contextValue}>
			<TooltipProviderGlobal>
				<ActionReleaseModal />
				<ServiceEditModal />
				{children}
			</TooltipProviderGlobal>
		</ModalProviderContext>
	);
};

export { ModalProviderContext, ModalProvider };
