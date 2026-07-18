import { ConfirmDestructiveDialog } from '@/components/confirm-dialog';
import { QUERY_KEYS } from '@/lib/query-keys';
import type { AuthUser } from '@/types/auth';
import * as authAPI from '@/utils/api/auth';

type DeleteUserDialogProps = {
	/** The user to delete; null keeps the dialog closed. */
	user: AuthUser | null;
	onOpenChange: (open: boolean) => void;
};

export const DeleteUserDialog = ({
	user,
	onOpenChange,
}: DeleteUserDialogProps) => (
	<ConfirmDestructiveDialog
		confirmLabel="Delete"
		description="Their sessions and API tokens are revoked immediately. This cannot be undone."
		mutationFn={(target: AuthUser) => authAPI.deleteUser(target.id)}
		onOpenChange={onOpenChange}
		queryKey={QUERY_KEYS.AUTH.USERS()}
		successMessage={(target) => `Deleted user '${target.username}'`}
		target={user}
		title={`Delete user '${user?.username}'?`}
	/>
);
