import { useQuery } from '@tanstack/react-query';
import { Pencil, Trash2 } from 'lucide-react';
import { type ReactElement, useState } from 'react';
import { Button } from '@/components/ui/button';
import { TableCell, TableHead, TableRow } from '@/components/ui/table';
import { useAuth } from '@/contexts/auth';
import { QUERY_KEYS } from '@/lib/query-keys';
import { ListPage } from '@/components/list-page';
import { DeleteUserDialog } from '@/pages/admin/users/delete-dialog';
import { UserDialog } from '@/pages/admin/users/user-dialog';
import type { AuthUser } from '@/types/auth';
import * as authAPI from '@/utils/api/auth';

/**
 * The user administration page: list, create, edit, and delete users.
 */
export const Users = (): ReactElement => {
	const { user: currentUser } = useAuth();
	const [dialogUser, setDialogUser] = useState<AuthUser | null>(null);
	const [dialogOpen, setDialogOpen] = useState(false);
	const [deleting, setDeleting] = useState<AuthUser | null>(null);

	const users = useQuery({
		queryFn: authAPI.listUsers,
		queryKey: QUERY_KEYS.AUTH.USERS(),
	});
	const groups = useQuery({
		queryFn: authAPI.listGroups,
		queryKey: QUERY_KEYS.AUTH.GROUPS(),
	});

	const openCreate = () => {
		setDialogUser(null);
		setDialogOpen(true);
	};
	const openEdit = (user: AuthUser) => {
		setDialogUser(user);
		setDialogOpen(true);
	};

	return (
		<>
			<ListPage
				addLabel="Add user"
				columns={
					<>
						<TableHead>Username</TableHead>
						<TableHead>Display name</TableHead>
						<TableHead>Groups</TableHead>
						<TableHead>Enabled</TableHead>
						<TableHead className="text-right">Actions</TableHead>
					</>
				}
				error={users.error}
				errorResource="users"
				isError={users.isError}
				onAdd={openCreate}
				tableLabel="Users"
				title="Users"
			>
				{(users.data ?? []).map((user) => (
					<TableRow className="odd:bg-muted/30" key={user.id}>
						<TableCell>{user.username}</TableCell>
						<TableCell>{user.display_name}</TableCell>
						<TableCell>{(user.groups ?? []).join(', ')}</TableCell>
						<TableCell>{user.enabled ? 'Yes' : 'No'}</TableCell>
						<TableCell className="flex justify-end gap-1">
							<Button
								aria-label={`Edit user ${user.username}`}
								onClick={() => openEdit(user)}
								size="icon-md"
								variant="ghost"
							>
								<Pencil aria-hidden />
							</Button>
							{user.id !== currentUser?.id && (
								<Button
									aria-label={`Delete user ${user.username}`}
									onClick={() => setDeleting(user)}
									size="icon-md"
									variant="ghost"
								>
									<Trash2 aria-hidden />
								</Button>
							)}
						</TableCell>
					</TableRow>
				))}
			</ListPage>

			<UserDialog
				groups={groups.data ?? []}
				onOpenChange={setDialogOpen}
				open={dialogOpen}
				user={dialogUser}
			/>
			<DeleteUserDialog
				onOpenChange={(open) => {
					if (!open) setDeleting(null);
				}}
				user={deleting}
			/>
		</>
	);
};
