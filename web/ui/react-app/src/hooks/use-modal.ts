import { useCallback, useMemo, useState } from 'react';
import type { ServiceModal } from '@/utils/api/types/config/summary';

/**
 * @returns The modal and a function to handle the modal.
 */
const useModal = () => {
	const [modal, setModal] = useState<ServiceModal>({
		actionType: '',
		service: { id: '', loading: true },
	});

	const hideModal = useCallback(() => {
		setModal({ actionType: '', service: { id: '', loading: true } });
	}, []);

	return useMemo(() => ({ hideModal, modal, setModal }), [hideModal, modal]);
};

export default useModal;
