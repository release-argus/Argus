import { ConfirmDestructiveDialog } from '@/components/confirm-dialog';
import { QUERY_KEYS } from '@/lib/query-keys';
import type { AuthGroup } from '@/types/auth';
import * as authAPI from '@/utils/api/auth';

type DeleteGroupDialogProps = {
	/** The group to delete; null keeps the dialog closed. */
	group: AuthGroup | null;
	onOpenChange: (open: boolean) => void;
};

export const DeleteGroupDialog = ({
	group,
	onOpenChange,
}: DeleteGroupDialogProps) => (
	<ConfirmDestructiveDialog
		confirmLabel="Delete"
		description="Members lose the group's permissions immediately. This cannot be undone."
		mutationFn={(target: AuthGroup) => authAPI.deleteGroup(target.id)}
		onOpenChange={onOpenChange}
		queryKey={QUERY_KEYS.AUTH.GROUPS()}
		successMessage={(target) => `Deleted group '${target.name}'`}
		target={group}
		title={`Delete group '${group?.name}'?`}
	/>
);
