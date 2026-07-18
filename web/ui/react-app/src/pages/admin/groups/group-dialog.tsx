import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useEffect, useMemo } from 'react';
import { FormProvider, useWatch } from 'react-hook-form';
import { toast } from 'sonner';
import FieldLabelWithTooltip from '@/components/generic/field-label';
import { Button } from '@/components/ui/button';
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from '@/components/ui/dialog';
import { Field, FieldError } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { TextOrLoading } from '@/components/ui/loading-ellipsis';
import useZodForm from '@/hooks/use-zod-form';
import { QUERY_KEYS } from '@/lib/query-keys';
import { diffGrants } from '@/pages/admin/groups/grant-diff';
import { GrantEditor } from '@/pages/admin/groups/grant-editor';
import { type GroupFormValues, groupSchema } from '@/pages/admin/groups/schema';
import {
	ADMIN_GROUP,
	type AuthGroup,
	type ResourcePermissions,
} from '@/types/auth';
import * as authAPI from '@/utils/api/auth';
import { getErrorMessage } from '@/utils/errors';

type GroupDialogProps = {
	/** The group being edited, or null to create a new one. */
	group: AuthGroup | null;
	catalogue: ResourcePermissions[];
	open: boolean;
	onOpenChange: (open: boolean) => void;
};

const emptyForm: GroupFormValues = {
	description: '',
	name: '',
	permissions: [],
};

/** Create/edit dialog for a group, including its permission grants. */
export const GroupDialog = ({
	group,
	catalogue,
	open,
	onOpenChange,
}: GroupDialogProps) => {
	const queryClient = useQueryClient();
	const form = useZodForm({
		defaultValues: emptyForm,
		schema: groupSchema,
	});
	const { errors } = form.formState;

	const permissions = useWatch({ control: form.control, name: 'permissions' });
	const savedGrants = group ? (group.permissions ?? []) : undefined;
	const grantDiff = useMemo(
		() => diffGrants(permissions ?? [], savedGrants),
		[permissions, savedGrants],
	);

	// The admin group's grants are fixed server-side.
	// system groups can't be renamed.
	const grantsLocked = group?.system === true && group.name === ADMIN_GROUP;

	// Load the target's values whenever the dialog opens.
	// biome-ignore lint/correctness/useExhaustiveDependencies: form stable.
	useEffect(() => {
		if (!open) return;
		form.reset(
			group
				? {
						description: group.description ?? '',
						name: group.name,
						permissions: group.permissions ?? [],
					}
				: emptyForm,
		);
	}, [group, open]);

	const save = useMutation({
		mutationFn: (values: GroupFormValues) => {
			if (group) {
				return authAPI.updateGroup(group.id, {
					description: values.description,
					name: group.system ? undefined : values.name,
					permissions: grantsLocked ? undefined : values.permissions,
				});
			}
			return authAPI.createGroup({
				description: values.description || undefined,
				name: values.name,
				permissions: values.permissions,
			});
		},
		onError: (error) => toast.error(getErrorMessage(error)),
		onSuccess: (saved) => {
			toast.success(`Saved group '${saved.name}'`);
			void queryClient.invalidateQueries({
				queryKey: QUERY_KEYS.AUTH.GROUPS(),
			});
			onOpenChange(false);
		},
	});

	return (
		<Dialog onOpenChange={onOpenChange} open={open}>
			<DialogContent
				aria-describedby={undefined}
				className="overflow-x-hidden sm:max-w-3xl"
			>
				<DialogHeader>
					<DialogTitle>
						{group ? `Edit group '${group.name}'` : 'Add group'}
					</DialogTitle>
				</DialogHeader>
				<FormProvider {...form}>
					<form
						className="grid min-w-0 gap-4"
						onSubmit={form.handleSubmit((values) => save.mutate(values))}
					>
						<Field className="gap-2" data-invalid={!!errors.name}>
							<FieldLabelWithTooltip
								htmlFor="group-name"
								required
								size="sm"
								text="Name"
							/>
							<Input
								aria-describedby={errors.name ? 'group-name-error' : undefined}
								aria-invalid={!!errors.name}
								aria-required
								disabled={group?.system === true}
								id="group-name"
								{...form.register('name')}
							/>
							<FieldError errors={[errors.name]} id="group-name-error" />
						</Field>
						<Field className="gap-2">
							<FieldLabelWithTooltip
								htmlFor="group-description"
								size="sm"
								text="Description"
							/>
							<Input id="group-description" {...form.register('description')} />
						</Field>
						{grantsLocked ? (
							<p className="text-muted-foreground text-sm">
								The admin group always holds every permission.
							</p>
						) : catalogue.length === 0 ? (
							<p className="text-muted-foreground text-sm">
								Permissions are unavailable - reload the page to try again.
							</p>
						) : (
							<GrantEditor catalogue={catalogue} diff={grantDiff} />
						)}
						<DialogFooter>
							<Button
								disabled={
									save.isPending ||
									!form.formState.isDirty ||
									grantDiff.duplicateKeys.size > 0
								}
								type="submit"
							>
								<TextOrLoading loading={save.isPending} text="Save" />
							</Button>
						</DialogFooter>
					</form>
				</FormProvider>
			</DialogContent>
		</Dialog>
	);
};
