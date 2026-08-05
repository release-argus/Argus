import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { Controller, useWatch } from 'react-hook-form';
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
import { Field, FieldDescription, FieldError } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from '@/components/ui/select';
import useZodForm from '@/hooks/use-zod-form';
import { QUERY_KEYS } from '@/lib/query-keys';
import { cn } from '@/lib/utils';
import {
	EXPIRY_OPTIONS,
	isValidDuration,
	type TokenFormValues,
	tokenSchema,
} from '@/pages/account/tokens/schema';
import type { APITokenCreated } from '@/types/auth';
import * as authAPI from '@/utils/api/auth';
import { getErrorMessage } from '@/utils/errors';

const emptyForm: TokenFormValues = {
	customExpiry: '',
	expiry: 'never',
	name: '',
};

type CreateTokenDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** Hands the new token to the caller for its one-time reveal. */
	onCreated: (token: APITokenCreated) => void;
};

export const CreateTokenDialog = ({
	open,
	onOpenChange,
	onCreated,
}: CreateTokenDialogProps) => {
	const queryClient = useQueryClient();
	const form = useZodForm({ defaultValues: emptyForm, schema: tokenSchema });
	const { errors } = form.formState;

	// biome-ignore lint/correctness/useExhaustiveDependencies: form stable.
	useEffect(() => {
		if (open) form.reset(emptyForm);
	}, [open]);

	const {
		mutate: createToken,
		isPending,
		reset,
	} = useMutation({
		mutationFn: (values: TokenFormValues) =>
			authAPI.createToken({
				expires_in:
					values.expiry === 'never'
						? undefined
						: values.expiry === 'custom'
							? values.customExpiry.trim()
							: values.expiry,
				name: values.name,
			}),
		onError: (error) => toast.error(getErrorMessage(error)),
		onSuccess: (token) => {
			void queryClient.invalidateQueries({
				queryKey: QUERY_KEYS.AUTH.TOKENS(),
			});
			onOpenChange(false);
			onCreated(token);
			// Don't persist in cache.
			reset();
		},
	});

	const { name, expiry, customExpiry } = useWatch({ control: form.control });
	const isCustom = expiry === 'custom';
	const customValue = customExpiry ?? '';
	const customInvalid = isCustom && !isValidDuration(customValue);
	const customError =
		isCustom && customValue !== '' && !isValidDuration(customValue);

	return (
		<Dialog onOpenChange={onOpenChange} open={open}>
			<DialogContent aria-describedby={undefined}>
				<DialogHeader>
					<DialogTitle>Create API token</DialogTitle>
				</DialogHeader>
				<form
					className="grid gap-4"
					onSubmit={form.handleSubmit((values) => {
						if (customInvalid) return;
						createToken(values);
					})}
				>
					<Field className="gap-2" data-invalid={!!errors.name}>
						<FieldLabelWithTooltip
							htmlFor="token-name"
							required
							size="sm"
							text="Name"
						/>
						<Input
							aria-invalid={!!errors.name}
							aria-required
							id="token-name"
							placeholder="e.g. ci-server"
							{...form.register('name')}
						/>
						<FieldError errors={[errors.name]} />
					</Field>
					<Field className="gap-2" data-invalid={customError}>
						<FieldLabelWithTooltip
							htmlFor="token-expiry"
							size="sm"
							text="Expires"
						/>
						<div className={cn('grid gap-2', isCustom && 'sm:grid-cols-2')}>
							<Controller
								control={form.control}
								name="expiry"
								render={({ field }) => (
									<Select onValueChange={field.onChange} value={field.value}>
										<SelectTrigger
											aria-label="Token expiry"
											className="w-full"
											id="token-expiry"
										>
											<SelectValue />
										</SelectTrigger>
										<SelectContent>
											{EXPIRY_OPTIONS.map((option) => (
												<SelectItem key={option.value} value={option.value}>
													{option.label}
												</SelectItem>
											))}
										</SelectContent>
									</Select>
								)}
							/>
							{isCustom && (
								<Input
									aria-describedby="token-custom-expiry-hint"
									aria-invalid={customError}
									aria-label="Custom duration"
									aria-required
									id="token-custom-expiry"
									placeholder="e.g. 48h, 90m, 1h30m"
									{...form.register('customExpiry')}
								/>
							)}
						</div>
						{isCustom && (
							<FieldDescription
								className={cn(customError && 'text-destructive')}
								id="token-custom-expiry-hint"
							>
								A duration such as 48h, 90m, or 1h30m (units: h, m, s).
							</FieldDescription>
						)}
					</Field>
					<DialogFooter>
						<Button
							disabled={isPending || !name || customInvalid}
							type="submit"
						>
							Create
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
};
