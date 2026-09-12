import { useMutation, useQueryClient } from '@tanstack/react-query';
import { type ReactElement, useEffect } from 'react';
import { useWatch } from 'react-hook-form';
import { toast } from 'sonner';
import { PasswordMismatchError } from '@/components/auth/password-mismatch-error';
import FieldLabelWithTooltip from '@/components/generic/field-label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Field, FieldDescription, FieldError } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { TextOrLoading } from '@/components/ui/loading-ellipsis';
import { MaskedInput } from '@/components/ui/masked-input';
import { useAuth } from '@/contexts/auth';
import usePasswordMismatch from '@/hooks/use-password-mismatch';
import useZodForm from '@/hooks/use-zod-form';
import { QUERY_KEYS } from '@/lib/query-keys';
import { GroupHeading, GroupRule } from '@/pages/account/group';
import {
	type AccountFormValues,
	accountSchema,
} from '@/pages/account/profile/schema';
import type { AuthMe, AuthUser } from '@/types/auth';
import * as authAPI from '@/utils/api/auth';
import type { AccountUpdateRequest } from '@/utils/api/types/requests/auth';
import { getErrorMessage } from '@/utils/errors';

/** The API's message when the current password does not match. */
const WRONG_CURRENT_PASSWORD = 'current password is incorrect';

/** Form field of an API field error, keyed by the name the API reports. */
const FIELD_OF_KEY: Record<string, keyof AccountFormValues> = {
	display_name: 'display_name',
	email: 'email',
	password: 'password',
};

type ServerError = {
	field: keyof AccountFormValues | 'root';
	message: string;
};

/** Sentence-cases an API message, which arrives lower-case and unpunctuated. */
const sentence = (text: string) =>
	text
		? `${text[0].toUpperCase()}${text.slice(1)}${text.endsWith('.') ? '' : '.'}`
		: text;

/** Where an API error belongs, and how the form words it. */
const serverError = (message: string): ServerError => {
	const field =
		message === WRONG_CURRENT_PASSWORD
			? 'currentPassword'
			: FIELD_OF_KEY[message.split(':')[0]];
	if (!field) return { field: 'root', message: sentence(message) };

	// Field errors arrive as `key: "value" <invalid> (why)`; show just the why.
	const why = /\((.+)\)$/.exec(message)?.[1];
	return { field: field, message: sentence(why ?? message) };
};

const formFor = (user?: AuthUser): AccountFormValues => ({
	confirmPassword: '',
	currentPassword: '',
	display_name: user?.display_name ?? '',
	email: user?.email ?? '',
	password: '',
});

/**
 * The signed-in user's own account details. Every change is confirmed with
 * the current password; a blank new password leaves password unchanged.
 */
export const Account = (): ReactElement => {
	const { user } = useAuth();
	const queryClient = useQueryClient();
	const form = useZodForm({
		defaultValues: formFor(user),
		schema: accountSchema,
	});
	const { errors } = form.formState;

	// biome-ignore lint/correctness/useExhaustiveDependencies: form stable.
	useEffect(() => {
		form.reset(formFor(user));
	}, [user]);

	const { currentPassword, display_name, email, password } = useWatch({
		control: form.control,
	});
	const {
		mismatch: passwordsMismatch,
		show: showPasswordMismatch,
		describedBy: passwordDescribedBy,
		errorID,
		mismatchID,
	} = usePasswordMismatch(form, { active: !!password, idPrefix: 'account-' });

	// Against the values last loaded from the API, so reverting an edit counts
	// as no change. The password fields confirm rather than change the account,
	// so they are not what makes it savable.
	const { defaultValues } = form.formState;
	const changed =
		(display_name ?? '') !== (defaultValues?.display_name ?? '') ||
		(email ?? '') !== (defaultValues?.email ?? '') ||
		!!password;

	// Blur validates (useZodForm's default), so hold back 'Required.' until
	// there is actually something to save.
	const currentPasswordError = changed ? errors.currentPassword : undefined;

	const save = useMutation({
		mutationFn: (values: AccountFormValues) => {
			const patch: AccountUpdateRequest = {
				current_password: values.currentPassword,
			};
			if (values.display_name !== (defaultValues?.display_name ?? ''))
				patch.display_name = values.display_name;
			if (values.email !== (defaultValues?.email ?? ''))
				patch.email = values.email;
			if (values.password) patch.new_password = values.password;
			return authAPI.updateAccount(patch);
		},
		onError: (error) => {
			const { field, message } = serverError(getErrorMessage(error));
			form.setError(field, { message: message, type: 'server' });
			if (field !== 'root') form.setFocus(field);
		},
		onSuccess: (me: AuthMe) => {
			queryClient.setQueryData(QUERY_KEYS.AUTH.ME(), me);
			toast.success(
				password
					? 'Account updated - other sessions signed out'
					: 'Account updated',
			);
			form.reset(formFor(me.user));
		},
	});

	return (
		<form
			className="flex flex-col gap-6"
			onSubmit={form.handleSubmit((values) => {
				if (passwordsMismatch || save.isPending || !changed || !currentPassword)
					return;
				form.clearErrors('root');
				save.mutate(values);
			})}
		>
			<Card>
				<CardContent className="grid gap-6">
					<section className="grid gap-4">
						<GroupHeading
							description="How you are identified across Argus."
							title="Profile"
						/>
						<div className="grid gap-4 sm:grid-cols-2">
							<Field className="gap-2">
								<FieldLabelWithTooltip
									htmlFor="account-username"
									size="sm"
									text="Username"
								/>
								<Input
									disabled
									id="account-username"
									readOnly
									value={user?.username ?? ''}
								/>
								<FieldDescription>
									Usernames cannot be changed.
								</FieldDescription>
							</Field>
							<Field className="gap-2" data-invalid={!!errors.display_name}>
								<FieldLabelWithTooltip
									htmlFor="account-display-name"
									size="sm"
									text="Display name"
								/>
								<Input
									aria-describedby={
										errors.display_name
											? 'account-display-name-error'
											: undefined
									}
									aria-invalid={!!errors.display_name}
									id="account-display-name"
									{...form.register('display_name')}
								/>
								<FieldError
									className="min-h-5"
									errors={[errors.display_name]}
									id="account-display-name-error"
								/>
							</Field>
							<Field className="gap-2" data-invalid={!!errors.email}>
								<FieldLabelWithTooltip
									htmlFor="account-email"
									size="sm"
									text="Email"
								/>
								<Input
									aria-describedby={
										errors.email ? 'account-email-error' : undefined
									}
									aria-invalid={!!errors.email}
									id="account-email"
									type="email"
									{...form.register('email')}
								/>
								<FieldError
									className="min-h-5"
									errors={[errors.email]}
									id="account-email-error"
								/>
							</Field>
						</div>
					</section>

					<GroupRule />

					<section className="grid gap-4">
						<GroupHeading
							description="Leave blank to keep your current password. Setting a new one signs out your other sessions."
							title="Password"
						/>
						<div className="grid gap-4 sm:grid-cols-2">
							<Field
								className="gap-2"
								data-invalid={!!errors.password || showPasswordMismatch}
							>
								<FieldLabelWithTooltip
									htmlFor="account-new-password"
									size="sm"
									text="New password"
								/>
								<MaskedInput
									aria-describedby={passwordDescribedBy}
									aria-invalid={!!errors.password || showPasswordMismatch}
									autoComplete="new-password"
									id="account-new-password"
									valueLabel="new password"
									{...form.register('password')}
								/>
								<FieldError
									className="min-h-5"
									errors={[errors.password]}
									id={errorID}
								/>
							</Field>
							<Field className="gap-2" data-invalid={showPasswordMismatch}>
								<FieldLabelWithTooltip
									htmlFor="account-confirm-password"
									size="sm"
									text="Confirm new password"
								/>
								<MaskedInput
									aria-describedby={
										showPasswordMismatch ? mismatchID : undefined
									}
									aria-invalid={showPasswordMismatch}
									autoComplete="new-password"
									disabled={!password}
									id="account-confirm-password"
									valueLabel="new password confirmation"
									{...form.register('confirmPassword')}
								/>
								{showPasswordMismatch && (
									<PasswordMismatchError id={mismatchID} />
								)}
							</Field>
						</div>
					</section>

					<GroupRule />

					<section className="grid gap-4">
						<GroupHeading
							description="Your current password is required to save any change above."
							title="Confirm changes"
						/>
						<div className="grid gap-4 sm:grid-cols-2">
							<Field className="gap-2" data-invalid={!!currentPasswordError}>
								<FieldLabelWithTooltip
									htmlFor="account-current-password"
									required
									size="sm"
									text="Current password"
								/>
								<MaskedInput
									aria-describedby={
										currentPasswordError
											? 'account-current-password-error'
											: undefined
									}
									aria-invalid={!!currentPasswordError}
									aria-required
									autoComplete="current-password"
									id="account-current-password"
									valueLabel="current password"
									{...form.register('currentPassword')}
								/>
								<FieldError
									className="min-h-5"
									errors={[currentPasswordError]}
									id="account-current-password-error"
								/>
							</Field>
						</div>
						{errors.root && (
							<Alert aria-label="Account update error" variant="destructive">
								<AlertDescription>{errors.root.message}</AlertDescription>
							</Alert>
						)}
					</section>
				</CardContent>
			</Card>

			{/* Left, not right: the toaster occupies the bottom-right corner. */}
			{/* aria-disabled rather than disabled: disabling a focused button blurs
			    it, so the keyboard user would lose their place on every save. */}
			<div className="flex">
				<Button
					aria-disabled={save.isPending || !changed || !currentPassword}
					className="aria-disabled:pointer-events-none aria-disabled:opacity-50"
					type="submit"
				>
					<TextOrLoading loading={save.isPending} text="Save changes" />
				</Button>
			</div>
		</form>
	);
};
