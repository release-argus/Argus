import { useQuery } from '@tanstack/react-query';
import { LoaderCircle, Lock } from 'lucide-react';
import { type ReactElement, useEffect, useState } from 'react';
import { useWatch } from 'react-hook-form';
import { Navigate, useLocation, useNavigate } from 'react-router';
import { z } from 'zod';
import { AuthCard } from '@/components/auth/auth-card';
import FieldLabelWithTooltip from '@/components/generic/field-label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Field } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { MaskedInput } from '@/components/ui/masked-input';
import { useAuth } from '@/contexts/auth';
import useZodForm from '@/hooks/use-zod-form';
import { QUERY_KEYS } from '@/lib/query-keys';
import { FirstRunSetup } from '@/pages/login/setup';
import * as authAPI from '@/utils/api/auth';
import { getErrorMessage } from '@/utils/errors';

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
	const [error, setError] = useState<string>();

	const { data: setupState } = useQuery({
		enabled: status === 'unauthenticated',
		queryFn: authAPI.fetchSetupState,
		queryKey: QUERY_KEYS.AUTH.SETUP(),
		retry: false,
	});
	const demo = setupState?.demo;

	const form = useZodForm({
		defaultValues: {
			password: '',
			username: '',
		},
		schema: loginSchema,
	});
	const { username, password } = useWatch({ control: form.control });

	// Demo credentials resolve after the first render, so fill the fields then.
	// biome-ignore lint/correctness/useExhaustiveDependencies: form stable.
	useEffect(() => {
		if (!demo) return;
		form.reset(
			{ password: demo.password, username: demo.username },
			{ keepDirtyValues: true },
		);
	}, [demo]);

	// Submitting, or still resolving whether a session already exists.
	const checkingSession = status === 'loading';
	const busy = form.formState.isSubmitting || checkingSession;
	const busyLabel = checkingSession ? 'Checking authentication' : 'Signing in';

	// Where to return to after logging in.
	const from =
		(location.state as { from?: string } | null)?.from ?? '/approvals';

	const onSubmit = form.handleSubmit(async (values) => {
		setError(undefined);
		try {
			await login(values.username, values.password);
			navigate(from, { replace: true });
		} catch (err) {
			setError(getErrorMessage(err));
		}
	});

	// Auth off, or already logged in: nothing to do here.
	if (status === 'auth-disabled' || status === 'authenticated') {
		return <Navigate replace to={from} />;
	}

	// No users yet: first-run setup replaces the login form.
	if (status === 'setup') {
		return <FirstRunSetup />;
	}

	return (
		<AuthCard icon={<Lock aria-hidden className="size-5" />} title="Sign in">
			<form aria-label="Login" className="grid gap-4" onSubmit={onSubmit}>
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
					<FieldLabelWithTooltip htmlFor="password" size="sm" text="Password" />
					<MaskedInput
						autoComplete="current-password"
						enterKeyHint="go"
						id="password"
						valueLabel="password"
						{...form.register('password')}
					/>
				</Field>
				{error && (
					<Alert aria-label="Login error" variant="destructive">
						<AlertDescription>{error}</AlertDescription>
					</Alert>
				)}
				<Button
					aria-disabled={!username || !password}
					aria-label={busy ? busyLabel : undefined}
					className="w-full aria-disabled:opacity-50"
					disabled={busy}
					type="submit"
				>
					{busy ? (
						<LoaderCircle aria-hidden className="animate-spin" />
					) : (
						'Sign in'
					)}
				</Button>
			</form>
		</AuthCard>
	);
};
