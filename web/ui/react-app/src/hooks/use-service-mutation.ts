import {
	type UseMutationOptions,
	useMutation,
	useQueryClient,
} from '@tanstack/react-query';
import type z from 'zod';
import { QUERY_KEYS } from '@/lib/query-keys';
import type { SuccessMessage } from '@/types/util';
import { mapRequest } from '@/utils/api/types/api-request-handler';

type UseServiceEditMutationProps<T extends z.ZodType> = {
	/* The ID of the service to edit. */
	serviceID: string | null;
	/* The data to send in the request. */
	data?: z.infer<T>;
};

type UseServiceMutationProps<T extends z.ZodType> = Pick<
	UseMutationOptions<SuccessMessage, unknown, UseServiceEditMutationProps<T>>,
	'onSuccess' | 'onError' | 'onSettled'
>;

/**
 * A custom hook to handle creating or updating a service.
 *
 * @template T - The Zod schema of the service form.
 * @param _schema - The Zod schema used for validation.
 * @param options - Optional callbacks for the mutation lifecycle:
 *   - `onError`: called if the mutation fails.
 *   - `onSuccess`: called if the mutation succeeds.
 *   - `onSettled`: called when the mutation is either successful or failed.
 * @returns The mutation object from `useMutation`, typed as:
 */
export const useServiceEdit = <T extends z.ZodType>(
	_schema: Omit<T, 'id_name_separator'>,
	{
		onError: onErrorProp,
		onSuccess: onSuccessProp,
		onSettled: onSettledProp,
	}: UseServiceMutationProps<T>,
) => {
	const queryClient = useQueryClient();

	return useMutation<SuccessMessage, unknown, UseServiceEditMutationProps<T>>({
		mutationFn: async ({ serviceID, data }) =>
			await mapRequest('SERVICE_EDIT', {
				body: data,
				serviceID: serviceID,
			}),
		onError: onErrorProp,
		onSettled: onSettledProp,
		onSuccess: (data, variables, onMutateResult, context) => {
			onSuccessProp?.(data, variables, onMutateResult, context);

			// Invalidate the service if edited.
			if (variables.serviceID) {
				void queryClient.invalidateQueries({
					queryKey: QUERY_KEYS.SERVICE.EDIT_ITEM(variables.serviceID),
					refetchType: 'none',
				});
			}
			// Invalidate service detail.
			void queryClient.invalidateQueries({
				queryKey: QUERY_KEYS.SERVICE.DETAIL(),
				refetchType: 'none',
			});
		},
	});
};

type UseServiceDeleteMutationProps = {
	/* The ID of the service to delete. */
	serviceID: string;
};

/**
 * A custom hook to handle deleting a service.
 *
 * @template T - The Zod schema of the service form.
 * @param _schema - The Zod schema used for validation.
 * @param options - Optional callbacks for the mutation lifecycle:
 *   - `onError`: called if the mutation fails.
 *   - `onSuccess`: called if the mutation succeeds.
 *   - `onSettled`: called when the mutation is either successful or failed.
 * @returns The mutation object from `useMutation`.
 */
export const useServiceDelete = <T extends z.ZodType>(
	_schema: T,
	{
		onError: onErrorProp,
		onSuccess: onSuccessProp,
		onSettled: onSettledProp,
	}: UseServiceMutationProps<T>,
) => {
	return useMutation<SuccessMessage, unknown, UseServiceDeleteMutationProps>({
		mutationFn: async ({ serviceID }) =>
			await mapRequest('SERVICE_DELETE', {
				serviceID: serviceID,
			}),
		onError: onErrorProp,
		onSettled: onSettledProp,
		onSuccess: onSuccessProp,
	});
};
