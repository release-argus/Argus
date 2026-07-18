import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { Controller, useWatch } from 'react-hook-form';
import { toast } from 'sonner';
import { PasswordMismatchError } from '@/components/auth/password-mismatch-error';
import FieldLabelWithTooltip from '@/components/generic/field-label';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from '@/components/ui/dialog';
import { Field, FieldDescription, FieldError } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { TextOrLoading } from '@/components/ui/loading-ellipsis';
import { MaskedInput } from '@/components/ui/masked-input';
import { useAuth } from '@/contexts/auth';
import usePasswordMismatch from '@/hooks/use-password-mismatch';
import useZodForm from '@/hooks/use-zod-form';
import { QUERY_KEYS } from '@/lib/query-keys';
import {
	type UserFormValues,
	userCreateSchema,
	userSchema,
} from '@/pages/admin/users/schema';
import { ADMIN_GROUP, type AuthGroup, type AuthUser } from '@/types/auth';
import * as authAPI from '@/utils/api/auth';
import type { UserPatchRequest } from '@/utils/api/types/requests/auth';
import { getErrorMessage } from '@/utils/errors';

const emptyForm: UserFormValues = {
	confirmPassword: '',
	display_name: '',
	email: '',
	enabled: true,
	groups: [],
	password: '',
	username: '',
};

type UserDialogProps = {
	/** The user being edited, or null to create a new one. */
	user: AuthUser | null;
	groups: AuthGroup[];
	open: boolean;
	onOpenChange: (open: boolean) => void;
};

/** Create/edit dialog for a user, including group membership. */
export const UserDialog = ({
	user,
	groups,
	open,
	onOpenChange,
}: UserDialogProps) => {
	const { user: currentUser } = useAuth();
	const queryClient = useQueryClient();
	const form = useZodForm({
		defaultValues: emptyForm,
		schema: user ? userSchema : userCreateSchema,
	});
	const { errors } = form.formState;

	// Cannot disable your own account,
	// nor remove it from the admin group,
	// nor delete it.
	const isSelf = user !== null && user.id === currentUser?.id;

	const { password } = useWatch({ control: form.control });
	const {
		mismatch: passwordsMismatch,
		show: showPasswordMismatch,
		describedBy: passwordDescribedBy,
		errorID,
		mismatchID,
	} = usePasswordMismatch(form, {
		active: isSelf && !!password,
		idPrefix: 'user-',
	});

	// biome-ignore lint/correctness/useExhaustiveDependencies: form stable.
	useEffect(() => {
		if (!open) return;
		form.reset(
			user
				? {
						confirmPassword: '',
						display_name: user.display_name ?? '',
						email: user.email ?? '',
						enabled: user.enabled,
						groups: user.groups ?? [],
						password: '',
						username: user.username,
					}
				: emptyForm,
		);
	}, [user, open]);

	const save = useMutation({
		mutationFn: (values: UserFormValues) => {
			if (user) {
				const patch: UserPatchRequest = {
					display_name: values.display_name,
					email: values.email,
					enabled: values.enabled,
					groups: values.groups,
				};
				if (values.password) patch.password = values.password;
				return authAPI.updateUser(user.id, patch);
			}
			return authAPI.createUser({
				display_name: values.display_name || undefined,
				email: values.email || undefined,
				groups: values.groups,
				password: values.password,
				username: values.username,
			});
		},
		onError: (error) => toast.error(getErrorMessage(error)),
		onSuccess: (saved) => {
			toast.success(`Saved user '${saved.username}'`);
			void queryClient.invalidateQueries({ queryKey: QUERY_KEYS.AUTH.USERS() });
			onOpenChange(false);
		},
	});

	return (
		<Dialog onOpenChange={onOpenChange} open={open}>
			<DialogContent aria-describedby={undefined}>
				<DialogHeader>
					<DialogTitle>
						{user ? `Edit user '${user.username}'` : 'Add user'}
					</DialogTitle>
				</DialogHeader>
				<form
					className="grid gap-4"
					onSubmit={form.handleSubmit((values) => {
						if (passwordsMismatch) return;
						save.mutate(values);
					})}
				>
					{!user && (
						<Field className="gap-2" data-invalid={!!errors.username}>
							<FieldLabelWithTooltip
								htmlFor="user-username"
								required
								size="sm"
								text="Username"
							/>
							<Input
								aria-describedby={
									errors.username ? 'user-username-error' : undefined
								}
								aria-invalid={!!errors.username}
								aria-required
								autoCapitalize="none"
								id="user-username"
								{...form.register('username')}
							/>
							<FieldError errors={[errors.username]} id="user-username-error" />
						</Field>
					)}
					<Field
						className="gap-2"
						data-invalid={!!errors.password || showPasswordMismatch}
					>
						<FieldLabelWithTooltip
							htmlFor="user-password"
							required={!user}
							size="sm"
							text={user ? 'Password (leave blank to keep)' : 'Password'}
						/>
						<MaskedInput
							aria-describedby={passwordDescribedBy}
							aria-invalid={!!errors.password || showPasswordMismatch}
							aria-required={!user}
							autoComplete="new-password"
							id="user-password"
							valueLabel="password"
							{...form.register('password')}
						/>
						<FieldError errors={[errors.password]} id={errorID} />
					</Field>
					{isSelf && (
						<Field className="gap-2" data-invalid={showPasswordMismatch}>
							<FieldLabelWithTooltip
								htmlFor="user-confirm-password"
								size="sm"
								text="Confirm password"
							/>
							<MaskedInput
								aria-describedby={showPasswordMismatch ? mismatchID : undefined}
								aria-invalid={showPasswordMismatch}
								autoComplete="new-password"
								id="user-confirm-password"
								valueLabel="password confirmation"
								{...form.register('confirmPassword')}
							/>
							{showPasswordMismatch && (
								<PasswordMismatchError id={mismatchID} />
							)}
						</Field>
					)}
					<Field className="gap-2">
						<FieldLabelWithTooltip
							htmlFor="user-display-name"
							size="sm"
							text="Display name"
						/>
						<Input id="user-display-name" {...form.register('display_name')} />
					</Field>
					<Field className="gap-2">
						<FieldLabelWithTooltip
							htmlFor="user-email"
							size="sm"
							text="Email"
						/>
						<Input id="user-email" type="email" {...form.register('email')} />
					</Field>
					{user && (
						<Controller
							control={form.control}
							name="enabled"
							render={({ field }) => (
								<Field className="gap-2">
									<div className="flex items-center gap-2">
										<Checkbox
											checked={field.value}
											disabled={isSelf}
											id="user-enabled"
											onCheckedChange={(checked) =>
												field.onChange(checked === true)
											}
										/>
										<Label htmlFor="user-enabled">Enabled</Label>
									</div>
									{isSelf && (
										<FieldDescription>
											You cannot disable your own account.
										</FieldDescription>
									)}
								</Field>
							)}
						/>
					)}
					<Controller
						control={form.control}
						name="groups"
						render={({ field }) => (
							<fieldset className="grid gap-2">
								<legend className="pb-1 font-medium text-sm">Groups</legend>
								{groups.map((group) => {
									const checked = field.value.includes(group.name);
									// Self cannot drop out of the admin group.
									const locked =
										isSelf && group.name === ADMIN_GROUP && checked;
									return (
										<div className="flex items-center gap-2" key={group.id}>
											<Checkbox
												checked={checked}
												disabled={locked}
												id={`user-group-${group.name}`}
												onCheckedChange={(next) =>
													field.onChange(
														next === true
															? [...field.value, group.name]
															: field.value.filter(
																	(name) => name !== group.name,
																),
													)
												}
											/>
											<Label htmlFor={`user-group-${group.name}`}>
												{group.name}
											</Label>
										</div>
									);
								})}
								{isSelf && field.value.includes(ADMIN_GROUP) && (
									<FieldDescription>
										You cannot remove yourself from the admin group.
									</FieldDescription>
								)}
							</fieldset>
						)}
					/>
					<DialogFooter>
						<Button disabled={save.isPending} type="submit">
							<TextOrLoading loading={save.isPending} text="Save" />
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
};
