import type { ReactElement } from 'react';
import { FieldError } from '@/components/ui/field';

type PasswordMismatchErrorProps = {
	/** ID the password inputs point at with aria-describedby. */
	id: string;
};

/** Shown under the confirmation when a password and its confirmation differ. */
export const PasswordMismatchError = ({
	id,
}: PasswordMismatchErrorProps): ReactElement => (
	<FieldError errors={[{ message: 'Passwords do not match.' }]} id={id} />
);
