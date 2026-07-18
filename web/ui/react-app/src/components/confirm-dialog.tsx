import {
	type QueryKey,
	useMutation,
	useQueryClient,
} from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { toast } from 'sonner';
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { getErrorMessage } from '@/utils/errors';

type ConfirmDestructiveDialogProps<T> = {
	/** The item to act on; null keeps the dialog closed. */
	target: T | null;
	onOpenChange: (open: boolean) => void;
	title: ReactNode;
	description: ReactNode;
	/** Label of the confirming button, e.g. 'Delete'. */
	confirmLabel: string;
	mutationFn: (target: T) => Promise<unknown>;
	/** Invalidated once the mutation succeeds. */
	queryKey: QueryKey;
	successMessage: (target: T) => string;
};

/**
 * A confirm-then-mutate dialog for destructive actions.
 * Shared by the user, group and token pages.
 */
export const ConfirmDestructiveDialog = <T,>({
	confirmLabel,
	description,
	mutationFn,
	onOpenChange,
	queryKey,
	successMessage,
	target,
	title,
}: ConfirmDestructiveDialogProps<T>) => {
	const queryClient = useQueryClient();
	const act = useMutation({
		mutationFn,
		onError: (error) => toast.error(getErrorMessage(error)),
		onSuccess: (_, actedOn) => {
			toast.success(successMessage(actedOn));
			void queryClient.invalidateQueries({ queryKey });
			onOpenChange(false);
		},
	});

	return (
		<AlertDialog onOpenChange={onOpenChange} open={target !== null}>
			<AlertDialogContent>
				<AlertDialogHeader>
					<AlertDialogTitle>{title}</AlertDialogTitle>
					<AlertDialogDescription>{description}</AlertDialogDescription>
				</AlertDialogHeader>
				<AlertDialogFooter>
					<AlertDialogCancel>Cancel</AlertDialogCancel>
					<AlertDialogAction
						onClick={() => target !== null && act.mutate(target)}
					>
						{confirmLabel}
					</AlertDialogAction>
				</AlertDialogFooter>
			</AlertDialogContent>
		</AlertDialog>
	);
};
