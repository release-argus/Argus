import { Check, Copy } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { toast } from 'sonner';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import type { APITokenCreated } from '@/types/auth';

type RevealTokenDialogProps = {
	/** The freshly created token; null keeps the dialog closed. */
	token: APITokenCreated | null;
	onOpenChange: (open: boolean) => void;
};

/** Shows a new token's plaintext once, at creation. */
export const RevealTokenDialog = ({
	token,
	onOpenChange,
}: RevealTokenDialogProps) => {
	const [copied, setCopied] = useState(false);
	const inputRef = useRef<HTMLInputElement>(null);

	// Reset the copied state each time a new token is revealed.
	useEffect(() => {
		if (token) setCopied(false);
	}, [token]);

	const copy = () => {
		if (!token) return;
		if (!navigator.clipboard) {
			inputRef.current?.select();
			toast.info('Clipboard unavailable - copy the selected token manually');
			return;
		}
		void navigator.clipboard
			.writeText(token.token)
			.then(() => {
				setCopied(true);
				toast.success('Token copied to clipboard');
			})
			.catch(() => {
				inputRef.current?.select();
				toast.error('Copy failed - copy the selected token manually');
			});
	};

	return (
		<Dialog onOpenChange={onOpenChange} open={token !== null}>
			<DialogContent aria-describedby={undefined}>
				<DialogHeader>
					<DialogTitle>Token '{token?.name}' created</DialogTitle>
				</DialogHeader>
				<Alert>
					<AlertDescription>
						Copy the token now - it will not be shown again.
					</AlertDescription>
				</Alert>
				<div className="flex items-center gap-2">
					<Input
						aria-label="New token"
						className="font-mono"
						readOnly
						ref={inputRef}
						value={token?.token ?? ''}
					/>
					<Button
						aria-label="Copy token"
						onClick={copy}
						size="icon-md"
						variant="outline"
					>
						{copied ? <Check aria-hidden /> : <Copy aria-hidden />}
					</Button>
				</div>
				<DialogFooter>
					<Button onClick={() => onOpenChange(false)} variant="secondary">
						Done
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
};
