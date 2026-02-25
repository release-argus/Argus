import { useMutation } from '@tanstack/react-query';
import { CircleCheck, CircleX, LoaderCircle, RotateCw } from 'lucide-react';
import { type FC, useMemo, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import { useErrors } from '@/hooks/use-error';
import { QUERY_KEYS } from '@/lib/query-keys';
import { beautifyGoErrors } from '@/utils';
import { mapRequest } from '@/utils/api/types/api-request-handler';
import type { NotifySchemaValues } from '@/utils/api/types/config-edit/notify/schemas';

/* The throttle time for sending the test message. */
const TEST_THROTTLE_MS = 2000;

/**
 * @param status - The status of the test message request.
 * @param msg - The response from the test message request.
 * @returns The result of the test message.
 */
const Result: FC<{
	status: ReturnType<typeof useMutation>['status'];
	msg?: string;
}> = ({ status, msg }) => {
	if (status === 'pending' || status === 'idle') return null;
	return (
		<span className="mb-2 w-100">
			<Alert variant={status === 'error' ? 'destructive' : 'default'}>
				<AlertTitle className="sr-only">Error sending:</AlertTitle>
				<AlertDescription className="whitespace-pre-wrap">
					{status === 'error' ? beautifyGoErrors(msg ?? 'error') : 'Success!'}
				</AlertDescription>
			</Alert>
		</span>
	);
};

type TestNotifyProps = {
	/* The path to the notify field. */
	path: string;
	/* The original notify field values. */
	original?: NotifySchemaValues;
	/* Extra data to send with the notify request. */
	extras: {
		service_url?: string;
		web_url?: string;
	};
};

/**
 * The `notify` form fields for testing.
 *
 * @param path - The path to the notify field.
 * @param original - The original notify field values.
 * @param extras - Extra data to send with the notify request.
 * @returns The `notify` form fields for testing.
 */
const TestNotify: FC<TestNotifyProps> = ({ path, original, extras }) => {
	const { getValues, trigger } = useFormContext();
	const {
		serviceID: serviceIDPrevious,
		schema: schemaFull,
		typeDataDefaults,
	} = useSchemaContext();

	const [lastFetched, setLastFetched] = useState(0);
	const errors = useErrors(path, true);

	const schema = schemaFull.shape.notify;

	const {
		data: testData,
		error: testError,
		isPending,
		mutateAsync: testRefetch,
		status: testStatus,
	} = useMutation({
		mutationFn: async () => {
			const result = schema.safeParse([getValues(path)]);
			if (!result.success) return null;

			const notifyType = result.data[0].type;
			return await mapRequest('NOTIFY_TEST', {
				defaults: typeDataDefaults?.notify?.[notifyType],
				extras: extras,
				new: result.data[0],
				previous: original,
				previousServiceID: serviceIDPrevious,
				serviceID: getValues('id') as string,
				serviceName: getValues('name') as string,
				type: notifyType,
			});
		},
		mutationKey: QUERY_KEYS.NOTIFY.TEST(serviceIDPrevious, original?.name),
		retry: false,
	});

	// Send the test notify request.
	const refetch = async () => {
		// Prevent refetching too often.
		const currentTime = Date.now();
		if (currentTime - lastFetched < TEST_THROTTLE_MS) return;

		try {
			const result = await trigger(path, { shouldFocus: true });
			if (result) {
				setLastFetched(currentTime);
				await testRefetch();
			}
		} catch (error) {}
	};

	// Icon for the test result.
	const resultIcon = useMemo(() => {
		if (isPending) return <LoaderCircle className="ml-2 animate-spin" />;
		if (testStatus !== 'idle') {
			return testStatus === 'error' ? (
				<CircleX className="text-destructive" />
			) : (
				<CircleCheck className="text-success" />
			);
		}
		return <></>;
	}, [isPending, testStatus]);

	return (
		<div className="col-span-full">
			<div className="ml-auto flex w-full flex-row items-center pb-2">
				{resultIcon}
				<Button
					className="ml-auto flex gap-1 px-2"
					disabled={isPending}
					onClick={refetch}
					variant="secondary"
				>
					<RotateCw />
					Send Test Message
				</Button>
			</div>
			{/* Render either the server error or form validation error */}
			<Result
				msg={testData?.message ?? testError?.message}
				status={testStatus}
			/>
			{errors && (
				<Alert variant="destructive">
					<AlertTitle className="sr-only">Errors in notify:</AlertTitle>
					<AlertDescription>
						{Object.entries(errors).map(([key, error]) => (
							<li key={key}>
								{key}: {error}
							</li>
						))}
					</AlertDescription>
				</Alert>
			)}
		</div>
	);
};

export default TestNotify;
