import { Pencil, Plus, Save } from 'lucide-react';
import type { FC } from 'react';
import { useToolbar } from '@/components/approvals/toolbar/toolbar-context';
import { Button } from '@/components/ui/button';
import Tip from '@/components/ui/tip';
import { useAuth } from '@/contexts/auth';
import useModal from '@/hooks/use-modal';

/**
 * Toolbar control for edit mode: toggles it, and while active shows buttons to
 * create a service and save order changes.
 */
const EditModeToggle: FC = () => {
	const { values, toggleEditMode, onSaveOrder, hasOrderChanged } = useToolbar();
	const { setModal } = useModal();
	const { hasPermission, hasAnyPermission } = useAuth();

	const canCreate = hasPermission('service', 'create');
	const canOrder = hasPermission('service_order', 'update');
	const canEditSome = hasAnyPermission('service', 'update');
	if (!canCreate && !canOrder && !canEditSome) return null;

	return (
		<>
			{values.editMode && (
				<>
					{canCreate && (
						<Tip
							content="Create a service"
							delayDuration={500}
							touchDelayDuration={250}
						>
							<Button
								aria-label="Create a service"
								className="rounded-none"
								id="create-service"
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
					)}
					{canOrder && hasOrderChanged && (
						<Tip
							content="Save order"
							delayDuration={500}
							touchDelayDuration={250}
						>
							<Button
								aria-label="Save order"
								className="rounded-none"
								id="save-order"
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
					aria-label="Toggle edit mode"
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
