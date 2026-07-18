import { LoaderCircle, Lock } from 'lucide-react';
import { type ReactElement, useState } from 'react';
import { useWatch } from 'react-hook-form';
import {
	Navigate,
	useLocation,
	useNavigate,
	useSearchParams,
} from 'react-router';
import { z } from 'zod';
import FieldLabelWithTooltip from '@/components/generic/field-label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Field } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { useAuth } from '@/contexts/auth';
import useZodForm from '@/hooks/use-zod-form';
import { FirstRunSetup } from '@/pages/login/setup';
import { getErrorMessage } from '@/utils/errors';
import getBasename from '@/utils/get-basename';

const loginSchema = z.object({
	password: z.string().min(1),
	username: z.string().min(1),
});

/**
 * The login page. Users land here on first visit, after logout, and whenever
 * a session expires mid-use (any 401 redirects here); a successful login
 * returns them to the page they came from.
 */
export const Login = (): ReactElement => {
	const { status, login } = useAuth();
	const navigate = useNavigate();
	const location = useLocation();
	const [searchParams] = useSearchParams();
	const [error, setError] = useState<string>();

	const form = useZodForm({
		defaultValues: {
			password: searchParams.get('password') ?? '',
			username: searchParams.get('username') ?? '',
		},
		schema: loginSchema,
	});
	const { username, password } = useWatch({ control: form.control });

	// Where to return to after logging in.
	const from =
		(location.state as { from?: string } | null)?.from ?? '/approvals';

	// Auth off, or already logged in: nothing to do here.
	if (status === 'disabled' || status === 'authenticated') {
		return <Navigate replace to={from} />;
	}

	// No users yet: first-run setup replaces the login form.
	if (status === 'setup') {
		return <FirstRunSetup />;
	}

	const onSubmit = form.handleSubmit(async (values) => {
		setError(undefined);
		try {
			await login(values.username, values.password);
			navigate(from, { replace: true });
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
						<Lock aria-hidden className="size-5" />
						Sign in
					</CardTitle>
				</CardHeader>
				<CardContent>
					<form aria-label="Login" className="grid gap-4" onSubmit={onSubmit}>
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
								htmlFor="password"
								size="sm"
								text="Password"
							/>
							<Input
								autoComplete="current-password"
								id="password"
								type="password"
								{...form.register('password')}
							/>
						</Field>
						{error && (
							<Alert aria-label="Login error" variant="destructive">
								<AlertDescription>{error}</AlertDescription>
							</Alert>
						)}
						<Button
							aria-label={
								form.formState.isSubmitting ? 'Signing in' : undefined
							}
							className="w-full"
							disabled={form.formState.isSubmitting || !username || !password}
							type="submit"
						>
							{form.formState.isSubmitting ? (
								<LoaderCircle aria-hidden className="animate-spin" />
							) : (
								'Sign in'
							)}
						</Button>
					</form>
				</CardContent>
			</Card>
		</div>
	);
};
