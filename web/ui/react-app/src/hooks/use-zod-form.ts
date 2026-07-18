import { zodResolver } from '@hookform/resolvers/zod';
import {
	type Resolver,
	type UseFormProps,
	type UseFormReturn,
	useForm,
} from 'react-hook-form';
import type { ZodObject, z } from 'zod';

/**
 * `useForm` wired to a Zod schema for validation. Validates on blur by
 * default; pass `mode` to override.
 *
 * @param props - `useForm` props (minus `resolver`) plus the Zod `schema`.
 */
const useZodForm = <T extends ZodObject>(
	props: Omit<UseFormProps<z.infer<T>>, 'resolver'> & {
		schema: T;
	},
): UseFormReturn<z.infer<T>> => {
	const { schema, ...rest } = props;

	return useForm<z.infer<T>>({
		mode: 'onBlur',
		...rest,
		resolver: zodResolver(schema) as unknown as Resolver<z.infer<T>>,
	});
};

export default useZodForm;
