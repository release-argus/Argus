import { zodResolver } from '@hookform/resolvers/zod';
import { type Resolver, type UseFormProps, useForm } from 'react-hook-form';
import type { ZodObject, z } from 'zod';

/**
 * Wrapper around `useForm` that uses a Zod schema for validation.
 *
 * @param props - The props to pass to `useForm`.
 * @param props.schema - The Zod schema to use for validation.
 * @param props.resolverContext - Optional context passed to the Zod resolver. Accessible inside refinements via `ctx.common.context`.
 */
const useServiceForm = <T extends ZodObject>(
	props: Omit<UseFormProps<z.infer<T>>, 'resolver'> & {
		schema: T;
	},
) => {
	const { schema, ...rest } = props;

	return useForm<z.infer<T>>({
		...rest,
		mode: 'onBlur',
		resolver: zodResolver(schema) as Resolver<z.infer<T>>,
	});
};

export default useServiceForm;
