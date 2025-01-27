import { FC, useContext } from 'react';
import { faPen, faPlus } from '@fortawesome/free-solid-svg-icons';

import ButtonWithTooltip from 'components/generic/button-with-tooltip';
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
				<ButtonWithTooltip
					hoverTooltip
					tooltip="Create a service"
					onClick={() => handleModal('EDIT', { id: '', loading: false })}
					icon={<FontAwesomeIcon icon={faPlus} />}
				/>
			)}
			<ButtonWithTooltip
				hoverTooltip
				tooltip="Toggle edit mode"
				onClick={toggleEditMode}
				icon={<FontAwesomeIcon icon={faPen} />}
			/>
		</>
	);
};

export default EditModeToggle;
