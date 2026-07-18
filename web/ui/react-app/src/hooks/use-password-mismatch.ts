import { type UseFormReturn, useWatch } from 'react-hook-form';

/** The fields a form must carry to be checked for a password mismatch. */
type PasswordFields = {
	password: string;
	confirmPassword: string;
};

type PasswordMismatchOptions = {
	/** Whether a confirmation is required at all. Defaults to true */
	active?: boolean;
	/** Prefix for the field IDs, where a form namespaces them. */
	idPrefix?: string;
};

type PasswordMismatch = {
	/** The passwords differ - block the submit. */
	mismatch: boolean;
	/** Show any alerts. */
	show: boolean;
	/** The password input's aria-describedby, or undefined when nothing shows. */
	describedBy?: string;
	/** ID of the password field's own error message. */
	errorID: string;
	/** ID of the mismatch alert. */
	mismatchID: string;
};

/**
 * Matches a form's password against its confirmation.
 *
 * Zod can't compare across fields without dropping the ZodObject type
 * [useZodForm] needs, so the match lives here. The error is held back until the
 * confirm field has been left or a submit tried, so it doesn't fire mid-typing.
 */
const usePasswordMismatch = <T extends PasswordFields>(
	form: UseFormReturn<T>,
	{ active = true, idPrefix = '' }: PasswordMismatchOptions = {},
): PasswordMismatch => {
	const { errors, isSubmitted, touchedFields } = form.formState;
	const values = useWatch({ control: form.control });
	const password = (values as Partial<PasswordFields>).password ?? '';
	const confirmPassword =
		(values as Partial<PasswordFields>).confirmPassword ?? '';

	const errorID = `${idPrefix}password-error`;
	const mismatchID = `${idPrefix}password-mismatch`;

	const mismatch = active && password !== confirmPassword;
	const show =
		mismatch &&
		((!!touchedFields.password && !!touchedFields.confirmPassword) ||
			isSubmitted);

	return {
		describedBy:
			[errors.password ? errorID : null, show ? mismatchID : null]
				.filter(Boolean)
				.join(' ') || undefined,
		errorID: errorID,
		mismatch: mismatch,
		mismatchID: mismatchID,
		show: show,
	};
};

export default usePasswordMismatch;
