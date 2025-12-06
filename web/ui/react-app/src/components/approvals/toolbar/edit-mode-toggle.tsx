import { Pencil, Plus, Save } from 'lucide-react';
import type { FC } from 'react';
import { useToolbar } from '@/components/approvals/toolbar/toolbar-context';
import { Button } from '@/components/ui/button';
import Tip from '@/components/ui/tip';
import useModal from '@/hooks/use-modal.ts';

/**
 * EditModeToggle
 *
 * Toolbar control for toggling edit mode. Displays buttons for creating services
 * and saving order changes when edit mode is active.
 */
const EditModeToggle: FC = () => {
	const { values, toggleEditMode, onSaveOrder, hasOrderChanged } = useToolbar();
	const { setModal } = useModal();

	return (
		<>
			{values.editMode && (
				<>
					<Tip
						content="Create a service"
						delayDuration={500}
						touchDelayDuration={250}
					>
						<Button
							className="rounded-none"
							onClick={() =>
								setModal({
									actionType: 'EDIT',
									service: { id: '', loading: false },
								})
							}
							type="button"
							variant="outline"
						>
							<Plus />
						</Button>
					</Tip>
					{hasOrderChanged && (
						<Tip
							content="Save order"
							delayDuration={500}
							touchDelayDuration={250}
						>
							<Button
								className="rounded-none"
								onClick={onSaveOrder}
								type="button"
								variant="outline"
							>
								<Save />
							</Button>
						</Tip>
					)}
				</>
			)}
			<Tip
				content="Toggle edit mode"
				delayDuration={500}
				touchDelayDuration={250}
			>
				<Button
					className="rounded-s-none"
					onClick={toggleEditMode}
					type="button"
					variant="outline"
				>
					<Pencil />
				</Button>
			</Tip>
		</>
	);
};

export default EditModeToggle;
