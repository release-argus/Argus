import { LoaderCircle, ShieldPlus } from 'lucide-react';
import { type ReactElement, useState } from 'react';
import { useWatch } from 'react-hook-form';
import { useNavigate } from 'react-router';
import { z } from 'zod';
import { AuthCard } from '@/components/auth/auth-card';
import { PasswordMismatchError } from '@/components/auth/password-mismatch-error';
import FieldLabelWithTooltip from '@/components/generic/field-label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Field, FieldError } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { MaskedInput } from '@/components/ui/masked-input';
import { useAuth } from '@/contexts/auth';
import usePasswordMismatch from '@/hooks/use-password-mismatch';
import useZodForm from '@/hooks/use-zod-form';
import { MIN_PASSWORD_LENGTH, PASSWORD_LENGTH_MESSAGE } from '@/types/auth';
import { getErrorMessage } from '@/utils/errors';

const setupSchema = z.object({
	confirmPassword: z.string().min(1),
	displayName: z.string(),
	password: z.string().min(MIN_PASSWORD_LENGTH, PASSWORD_LENGTH_MESSAGE),
	username: z.string().min(1),
});

/**
 * The first-run setup form - shown (via the login page) while no users exist.
 * Creates the first administrator account.
 */
export const FirstRunSetup = (): ReactElement => {
	const { setup } = useAuth();
	const navigate = useNavigate();
	const [error, setError] = useState<string>();

	const form = useZodForm({
		defaultValues: {
			confirmPassword: '',
			displayName: '',
			password: '',
			username: '',
		},
		schema: setupSchema,
	});
	const { errors, isSubmitting } = form.formState;
	const { username, password, confirmPassword } = useWatch({
		control: form.control,
	});

	const {
		mismatch: passwordsMismatch,
		show: showPasswordMismatch,
		describedBy: passwordDescribedBy,
		mismatchID,
	} = usePasswordMismatch(form);

	const onSubmit = form.handleSubmit(async (values) => {
		if (passwordsMismatch) return;
		setError(undefined);
		try {
			await setup(values.username, values.displayName, values.password);
			navigate('/approvals', { replace: true });
		} catch (err) {
			setError(getErrorMessage(err));
		}
	});

	return (
		<AuthCard
			description="Create the administrator account to get started."
			icon={<ShieldPlus aria-hidden className="size-5" />}
			title="Welcome"
		>
			<form
				aria-label="First-run setup"
				className="grid gap-4"
				onSubmit={onSubmit}
			>
				<Field className="gap-2">
					<FieldLabelWithTooltip htmlFor="username" size="sm" text="Username" />
					<Input
						autoCapitalize="none"
						autoComplete="username"
						autoCorrect="off"
						autoFocus
						enterKeyHint="next"
						id="username"
						spellCheck={false}
						{...form.register('username')}
					/>
				</Field>
				<Field className="gap-2">
					<FieldLabelWithTooltip
						htmlFor="display-name"
						size="sm"
						text="Display name (optional)"
					/>
					<Input
						autoComplete="name"
						enterKeyHint="next"
						id="display-name"
						{...form.register('displayName')}
					/>
				</Field>
				<Field
					className="gap-2"
					data-invalid={!!errors.password || showPasswordMismatch}
				>
					<FieldLabelWithTooltip htmlFor="password" size="sm" text="Password" />
					<MaskedInput
						aria-describedby={passwordDescribedBy}
						aria-invalid={!!errors.password || showPasswordMismatch}
						enterKeyHint="next"
						id="password"
						valueLabel="password"
						{...form.register('password')}
					/>
					<FieldError errors={[errors.password]} id="password-error" />
				</Field>
				<Field className="gap-2" data-invalid={showPasswordMismatch}>
					<FieldLabelWithTooltip
						htmlFor="confirm-password"
						size="sm"
						text="Confirm password"
					/>
					<MaskedInput
						aria-describedby={showPasswordMismatch ? mismatchID : undefined}
						aria-invalid={showPasswordMismatch}
						enterKeyHint="go"
						id="confirm-password"
						valueLabel="password confirmation"
						{...form.register('confirmPassword')}
					/>
					{showPasswordMismatch && <PasswordMismatchError id={mismatchID} />}
				</Field>
				{error && (
					<Alert aria-label="Setup error" variant="destructive">
						<AlertDescription>{error}</AlertDescription>
					</Alert>
				)}
				<Button
					aria-disabled={!username || !password || !confirmPassword}
					aria-label={isSubmitting ? 'Creating administrator' : undefined}
					className="w-full aria-disabled:opacity-50"
					disabled={isSubmitting}
					type="submit"
				>
					{isSubmitting ? (
						<LoaderCircle aria-hidden className="animate-spin" />
					) : (
						'Create administrator'
					)}
				</Button>
			</form>
		</AuthCard>
	);
};
