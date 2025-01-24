import { FC, useContext } from 'react';
import { faPen, faPlus } from '@fortawesome/free-solid-svg-icons';

import { Button } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { ModalContext } from 'contexts/modal';

type Props = {
	editMode: boolean;
	toggleEditMode: () => void;
};

const EditModeToggle: FC<Props> = ({ editMode, toggleEditMode }) => {
	const { handleModal } = useContext(ModalContext);

	return (
		<>
			{editMode && (
				<Button
					variant="secondary"
					className="border-0"
					onClick={() => handleModal('EDIT', { id: '', loading: false })}
				>
					<FontAwesomeIcon icon={faPlus} />
				</Button>
			)}
			<Button variant="secondary" className="border-0" onClick={toggleEditMode}>
				<FontAwesomeIcon icon={faPen} />
			</Button>
		</>
	);
};

export default EditModeToggle;
