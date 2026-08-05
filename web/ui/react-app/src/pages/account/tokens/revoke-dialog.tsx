import { ConfirmDestructiveDialog } from '@/components/confirm-dialog';
import { QUERY_KEYS } from '@/lib/query-keys';
import type { APIToken } from '@/types/auth';
import * as authAPI from '@/utils/api/auth';

type RevokeTokenDialogProps = {
	/** The token to revoke; null keeps the dialog closed. */
	token: APIToken | null;
	onOpenChange: (open: boolean) => void;
};

export const RevokeTokenDialog = ({
	token,
	onOpenChange,
}: RevokeTokenDialogProps) => (
	<ConfirmDestructiveDialog
		confirmLabel="Revoke"
		description="Clients using this token lose access immediately. This cannot be undone."
		mutationFn={(target: APIToken) => authAPI.deleteToken(target.id)}
		onOpenChange={onOpenChange}
		queryKey={QUERY_KEYS.AUTH.TOKENS()}
		successMessage={(target) => `Revoked token '${target.name}'`}
		target={token}
		title={`Revoke token '${token?.name}'?`}
	/>
);
