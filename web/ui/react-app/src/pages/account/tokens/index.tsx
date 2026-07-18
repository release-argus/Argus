import { useQuery } from '@tanstack/react-query';
import { format } from 'date-fns';
import { Trash2 } from 'lucide-react';
import { type ReactElement, useState } from 'react';
import { ListPage } from '@/components/list-page';
import { Button } from '@/components/ui/button';
import { TableCell, TableHead, TableRow } from '@/components/ui/table';
import { QUERY_KEYS } from '@/lib/query-keys';
import { CreateTokenDialog } from '@/pages/account/tokens/create-dialog';
import { RevealTokenDialog } from '@/pages/account/tokens/reveal-dialog';
import { RevokeTokenDialog } from '@/pages/account/tokens/revoke-dialog';
import type { APIToken, APITokenCreated } from '@/types/auth';
import * as authAPI from '@/utils/api/auth';

/** Formats an ISO timestamp as YYYY/MM/DD, HH:MM:SS in local time. */
const formatDate = (value?: string) =>
	value ? format(new Date(value), 'yyyy/MM/dd, HH:mm:ss') : '-';

/**
 * The API tokens page: create, list, and revoke the signed-in user's own
 * access tokens. The plaintext token is shown once, at creation.
 */
export const Tokens = (): ReactElement => {
	const [creating, setCreating] = useState(false);
	const [created, setCreated] = useState<APITokenCreated | null>(null);
	const [revoking, setRevoking] = useState<APIToken | null>(null);

	const tokens = useQuery({
		queryFn: authAPI.listTokens,
		queryKey: QUERY_KEYS.AUTH.TOKENS(),
	});

	return (
		<>
			<ListPage
				addLabel="Create token"
				columns={
					<>
						<TableHead>Name</TableHead>
						<TableHead>Token</TableHead>
						<TableHead>Created</TableHead>
						<TableHead>Expires</TableHead>
						<TableHead>Last used</TableHead>
						<TableHead className="text-right">Actions</TableHead>
					</>
				}
				description={
					<p className="pb-4 text-muted-foreground text-sm">
						Tokens authenticate headless clients with
						<code className="mx-1">Authorization: Bearer &lt;token&gt;</code>
						and carry your permissions. Revoking a token takes effect
						immediately.
					</p>
				}
				error={tokens.error}
				errorResource="API tokens"
				isError={tokens.isError}
				onAdd={() => setCreating(true)}
				tableLabel="API tokens"
				title="API Tokens"
			>
				{(tokens.data ?? []).map((token) => (
					<TableRow className="odd:bg-muted/30" key={token.id}>
						<TableCell>{token.name}</TableCell>
						<TableCell className="font-mono">{token.prefix}&hellip;</TableCell>
						<TableCell>{formatDate(token.created_at)}</TableCell>
						<TableCell>
							{token.expires_at ? formatDate(token.expires_at) : 'Never'}
						</TableCell>
						<TableCell>{formatDate(token.last_used_at)}</TableCell>
						<TableCell className="flex justify-end">
							<Button
								aria-label={`Revoke token ${token.name}`}
								onClick={() => setRevoking(token)}
								size="icon-md"
								variant="ghost"
							>
								<Trash2 aria-hidden />
							</Button>
						</TableCell>
					</TableRow>
				))}
			</ListPage>

			<CreateTokenDialog
				onCreated={setCreated}
				onOpenChange={setCreating}
				open={creating}
			/>
			<RevealTokenDialog
				onOpenChange={(open) => {
					if (!open) setCreated(null);
				}}
				token={created}
			/>
			<RevokeTokenDialog
				onOpenChange={(open) => {
					if (!open) setRevoking(null);
				}}
				token={revoking}
			/>
		</>
	);
};
