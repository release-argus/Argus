import { use } from 'react';
import { ModalProviderContext } from '@/contexts/modal';

/**
 * Access the modal context and functions to control modals.
 *
 * @returns An object with:
 *   - `modal`: The current modal state (`actionType`, `service`, etc.).
 *   - `setModal`: Function to update the modal state.
 *   - `hideModal`: Function to hide the modal.
 *
 * @throws Error if used outside of a `ModalProvider`.
 */
const useModal = () => {
	const context = use(ModalProviderContext);
	if (!context) {
		throw new Error('useModalContext must be used within a ModalProvider');
	}
	return context;
};

export default useModal;
