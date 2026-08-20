import { useQuery } from '@tanstack/react-query';
import { Pencil, Trash2 } from 'lucide-react';
import { type ReactElement, useState } from 'react';
import { ListPage } from '@/components/list-page';
import { Button } from '@/components/ui/button';
import { TableCell, TableHead, TableRow } from '@/components/ui/table';
import Tip from '@/components/ui/tip';
import { QUERY_KEYS } from '@/lib/query-keys';
import { DeleteGroupDialog } from '@/pages/admin/groups/delete-dialog';
import { GroupDialog } from '@/pages/admin/groups/group-dialog';
import { describePermission } from '@/pages/admin/groups/permissions';
import type { AuthGroup, Grant } from '@/types/auth';
import * as authAPI from '@/utils/api/auth';

/** Renders a grant like `service:read @ service/argus`. */
const grantLabel = (grant: Grant) =>
	`${grant.resource}:${grant.action}` +
	(grant.scope.type === 'global'
		? ''
		: ` @ ${grant.scope.type}/${grant.scope.ref ?? ''}`);

/** A grant count that reveals the full list on click. */
const PermissionsCell = ({ grants }: { grants: Grant[] }) => {
	if (grants.length === 0) {
		return <span className="text-muted-foreground">None</span>;
	}
	return (
		<Tip
			content={
				<ul className="grid gap-1 text-left text-xs">
					{grants.map((grant, index) => (
						// biome-ignore lint/suspicious/noArrayIndexKey: stable order with no duplicates.
						<li key={index}>
							<span className="font-mono">{grantLabel(grant)}</span>
							<span className="block text-muted-foreground">
								{describePermission(grant.resource, grant.action)}
							</span>
						</li>
					))}
				</ul>
			}
		>
			<span className="cursor-pointer underline decoration-dotted underline-offset-2">
				{grants.length} permission{grants.length === 1 ? '' : 's'}
			</span>
		</Tip>
	);
};

/**
 * The group administration page: list, create, edit, and
 * delete groups. System groups cannot be renamed or
 * deleted, and the admin group's grants are fixed.
 */
export const Groups = (): ReactElement => {
	const [dialogGroup, setDialogGroup] = useState<AuthGroup | null>(null);
	const [dialogOpen, setDialogOpen] = useState(false);
	const [deleting, setDeleting] = useState<AuthGroup | null>(null);

	const groups = useQuery({
		queryFn: authAPI.listGroups,
		queryKey: QUERY_KEYS.AUTH.GROUPS(),
	});
	const catalogue = useQuery({
		queryFn: authAPI.fetchPermissionCatalogue,
		queryKey: QUERY_KEYS.AUTH.PERMISSIONS(),
		staleTime: Infinity,
	});

	const openCreate = () => {
		setDialogGroup(null);
		setDialogOpen(true);
	};
	const openEdit = (group: AuthGroup) => {
		setDialogGroup(group);
		setDialogOpen(true);
	};

	return (
		<>
			<ListPage
				addLabel="Add group"
				columns={
					<>
						<TableHead>Name</TableHead>
						<TableHead>Description</TableHead>
						<TableHead>Members</TableHead>
						<TableHead>Permissions</TableHead>
						<TableHead className="text-right">Actions</TableHead>
					</>
				}
				error={groups.error}
				errorResource="groups"
				isError={groups.isError}
				onAdd={openCreate}
				tableLabel="Groups"
				title="Groups"
			>
				{(groups.data ?? []).map((group) => (
					<TableRow className="odd:bg-muted/30" key={group.id}>
						<TableCell>
							{group.name}
							{group.system && (
								<span className="ml-2 text-muted-foreground text-xs">
									system
								</span>
							)}
						</TableCell>
						<TableCell>{group.description}</TableCell>
						<TableCell>{group.members}</TableCell>
						<TableCell>
							<PermissionsCell grants={group.permissions ?? []} />
						</TableCell>
						<TableCell className="flex justify-end gap-1">
							<Button
								aria-label={`Edit group ${group.name}`}
								onClick={() => openEdit(group)}
								size="icon-md"
								variant="ghost"
							>
								<Pencil aria-hidden />
							</Button>
							{!group.system && (
								<Button
									aria-label={`Delete group ${group.name}`}
									onClick={() => setDeleting(group)}
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

			<GroupDialog
				catalogue={catalogue.data?.resources ?? []}
				group={dialogGroup}
				onOpenChange={setDialogOpen}
				open={dialogOpen}
			/>
			<DeleteGroupDialog
				group={deleting}
				onOpenChange={(open) => {
					if (!open) setDeleting(null);
				}}
			/>
		</>
	);
};
