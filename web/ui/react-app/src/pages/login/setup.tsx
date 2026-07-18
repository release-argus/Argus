import { Eye, EyeOff, LoaderCircle, ShieldPlus } from 'lucide-react';
import { type ReactElement, useState } from 'react';
import { useWatch } from 'react-hook-form';
import { useNavigate } from 'react-router';
import { z } from 'zod';
import FieldLabelWithTooltip from '@/components/generic/field-label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Field } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import {
	InputGroup,
	InputGroupAddon,
	InputGroupButton,
	InputGroupInput,
} from '@/components/ui/input-group';
import { useAuth } from '@/contexts/auth';
import useZodForm from '@/hooks/use-zod-form';
import { getErrorMessage } from '@/utils/errors';
import getBasename from '@/utils/get-basename';

const setupSchema = z.object({
	confirmPassword: z.string().min(1),
	displayName: z.string(),
	password: z.string().min(1),
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
	const [showPassword, setShowPassword] = useState(false);
	const [showConfirm, setShowConfirm] = useState(false);

	const form = useZodForm({
		defaultValues: {
			confirmPassword: '',
			displayName: '',
			password: '',
			username: '',
		},
		schema: setupSchema,
	});
	const { isSubmitting, isSubmitted, touchedFields } = form.formState;
	const { username, password, confirmPassword } = useWatch({
		control: form.control,
	});

	// The zod schema can't check equality across fields without dropping the
	// ZodObject type useZodForm needs, so match them here.
	const passwordsMismatch =
		confirmPassword !== '' && password !== confirmPassword;
	// Hold the error back until the confirm field is left or a submit is tried.
	const showPasswordMismatch =
		passwordsMismatch &&
		((touchedFields.password && touchedFields.confirmPassword) || isSubmitted);

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
		<div className="flex min-h-[80vh] w-full flex-col items-center justify-center gap-6 p-5">
			<img
				alt="Argus"
				className="size-14"
				src={`${getBasename()}/favicon.svg`}
			/>
			<Card className="w-full max-w-sm">
				<CardHeader>
					<CardTitle className="flex items-center gap-2 text-2xl">
						<ShieldPlus aria-hidden className="size-5" />
						Welcome to Argus
					</CardTitle>
					<p className="text-muted-foreground text-sm">
						Create the administrator account to get started.
					</p>
				</CardHeader>
				<CardContent>
					<form
						aria-label="First-run setup"
						className="grid gap-4"
						onSubmit={onSubmit}
					>
						<Field className="gap-2">
							<FieldLabelWithTooltip
								htmlFor="username"
								size="sm"
								text="Username"
							/>
							<Input
								autoCapitalize="none"
								autoComplete="username"
								autoFocus
								id="username"
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
								id="display-name"
								{...form.register('displayName')}
							/>
						</Field>
						<Field className="gap-2">
							<FieldLabelWithTooltip
								htmlFor="password"
								size="sm"
								text="Password"
							/>
							<InputGroup>
								<InputGroupInput
									aria-describedby={
										showPasswordMismatch ? 'password-mismatch' : undefined
									}
									aria-invalid={showPasswordMismatch}
									autoComplete="new-password"
									id="password"
									type={showPassword ? 'text' : 'password'}
									{...form.register('password')}
								/>
								<InputGroupAddon align="inline-end">
									<InputGroupButton
										aria-label={
											showPassword ? 'Hide password' : 'Show password'
										}
										onClick={() => setShowPassword((shown) => !shown)}
										size="icon-xs"
									>
										{showPassword ? (
											<EyeOff aria-hidden />
										) : (
											<Eye aria-hidden />
										)}
									</InputGroupButton>
								</InputGroupAddon>
							</InputGroup>
						</Field>
						<Field className="gap-2">
							<FieldLabelWithTooltip
								htmlFor="confirm-password"
								size="sm"
								text="Confirm password"
							/>
							<InputGroup>
								<InputGroupInput
									aria-describedby={
										showPasswordMismatch ? 'password-mismatch' : undefined
									}
									aria-invalid={showPasswordMismatch}
									autoComplete="new-password"
									id="confirm-password"
									type={showConfirm ? 'text' : 'password'}
									{...form.register('confirmPassword')}
								/>
								<InputGroupAddon align="inline-end">
									<InputGroupButton
										aria-label={showConfirm ? 'Hide password' : 'Show password'}
										onClick={() => setShowConfirm((shown) => !shown)}
										size="icon-xs"
									>
										{showConfirm ? <EyeOff aria-hidden /> : <Eye aria-hidden />}
									</InputGroupButton>
								</InputGroupAddon>
							</InputGroup>
						</Field>
						{showPasswordMismatch && (
							<Alert
								aria-label="Password mismatch"
								id="password-mismatch"
								variant="destructive"
							>
								<AlertDescription>Passwords do not match.</AlertDescription>
							</Alert>
						)}
						{error && (
							<Alert aria-label="Setup error" variant="destructive">
								<AlertDescription>{error}</AlertDescription>
							</Alert>
						)}
						<Button
							aria-label={isSubmitting ? 'Creating administrator' : undefined}
							className="w-full"
							disabled={
								isSubmitting || !username || !password || !confirmPassword
							}
							type="submit"
						>
							{isSubmitting ? (
								<LoaderCircle aria-hidden className="animate-spin" />
							) : (
								'Create administrator'
							)}
						</Button>
					</form>
				</CardContent>
			</Card>
		</div>
	);
};
