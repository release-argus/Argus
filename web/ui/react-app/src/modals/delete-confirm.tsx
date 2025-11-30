import { LoaderCircle } from 'lucide-react';
import { type FC, useCallback, useState } from 'react';
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
	AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import { useServiceDelete } from '@/hooks/use-service-mutation';

type DeleteModalProps = {
	disabled?: boolean;
};

/**
 * A delete confirmation modal.
 *
 * @param disabled - Disable the button.
 * @returns A delete confirmation modal.
 */
export const DeleteModal: FC<DeleteModalProps> = ({ disabled }) => {
	const { serviceID, schema } = useSchemaContext();
	const [open, setOpen] = useState(false);

	const { mutate, isPending } = useServiceDelete(schema, {
		onSettled: () => {
			setOpen(false);
		},
	});

	// biome-ignore lint/correctness/useExhaustiveDependencies: mutate stable.
	const onClick = useCallback(() => {
		if (serviceID) mutate({ serviceID });
	}, [serviceID]);

	return (
		<AlertDialog onOpenChange={setOpen} open={open}>
			<AlertDialogTrigger asChild>
				<Button disabled={disabled} variant="destructive">
					Delete
				</Button>
			</AlertDialogTrigger>
			<AlertDialogContent className="backdrop-blur-sm">
				<AlertDialogHeader>
					<AlertDialogTitle>Confirm Delete</AlertDialogTitle>
				</AlertDialogHeader>
				<AlertDialogDescription>
					Are you sure you want to delete this item?
					<br />
					This action cannot be undone.
					{isPending && <LoaderCircle className="ml-2 animate-spin" />}
				</AlertDialogDescription>
				<AlertDialogFooter>
					<AlertDialogCancel disabled={isPending}>Cancel</AlertDialogCancel>
					<AlertDialogAction disabled={isPending} onClick={onClick}>
						Delete
					</AlertDialogAction>
				</AlertDialogFooter>
			</AlertDialogContent>
		</AlertDialog>
	);
};
