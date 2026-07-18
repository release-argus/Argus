import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { FormProvider } from 'react-hook-form';
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
import useZodForm from '@/hooks/use-zod-form';
import { QUERY_KEYS } from '@/lib/query-keys';
import { GrantEditor } from '@/pages/admin/groups/grant-editor';
import { type GroupFormValues, groupSchema } from '@/pages/admin/groups/schema';
import type { AuthGroup, ResourcePermissions } from '@/types/auth';
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

	// The admin group's grants are fixed server-side.
	// system groups can't be renamed.
	const grantsLocked = group?.system === true && group.name === 'admin';

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
				className="max-h-[85vh] overflow-y-auto overflow-x-hidden sm:max-w-3xl"
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
								aria-invalid={!!errors.name}
								aria-required
								disabled={group?.system === true}
								id="group-name"
								{...form.register('name')}
							/>
							<FieldError errors={[errors.name]} />
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
						) : (
							<GrantEditor
								catalogue={catalogue}
								savedGrants={group ? (group.permissions ?? []) : undefined}
							/>
						)}
						<DialogFooter>
							<Button
								disabled={save.isPending || !form.formState.isDirty}
								type="submit"
							>
								Save
							</Button>
						</DialogFooter>
					</form>
				</FormProvider>
			</DialogContent>
		</Dialog>
	);
};
