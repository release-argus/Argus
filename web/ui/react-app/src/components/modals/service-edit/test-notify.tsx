import { Alert, Button } from 'react-bootstrap';
import { FC, useMemo, useState } from 'react';
import { beautifyGoErrors, fetchJSON } from 'utils';
import {
	faCheckCircle,
	faCircleXmark,
	faSpinner,
	faSync,
} from '@fortawesome/free-solid-svg-icons';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { NotifyTypesValues } from 'types/config';
import { convertNotifyToAPI } from 'components/modals/service-edit/util/ui-api-conversions';
import { deepDiff } from 'utils/query-params';
import { useErrors } from 'hooks/errors';
import { useFormContext } from 'react-hook-form';
import { useQuery } from '@tanstack/react-query';

const Result: FC<{ status: 'pending' | 'success' | 'error'; err?: string }> = ({
	status,
	err,
}) => {
	if (status === 'pending') return <></>;
	return (
		<span className="mb-2" style={{ width: '100%', wordBreak: 'break-all' }}>
			<Alert variant={err || status === 'error' ? 'danger' : 'success'}>
				{err || status === 'error'
					? beautifyGoErrors(err ?? 'error')
					: 'Success!'}
			</Alert>
		</span>
	);
};

interface Props {
	path: string;
	original?: NotifyTypesValues;
	extras?: {
		service_id_previous?: string;
		service_url?: string;
		web_url?: string;
	};
}

const TestNotify: FC<Props> = ({ path, original, extras }) => {
	const { getValues, trigger } = useFormContext();
	const [lastFetched, setLastFetched] = useState(0);
	const errors = useErrors(path, true);

	const fetchTestNotifyJSON = (dataJSON: NotifyTypesValues) =>
		fetchJSON<{ message: string }>({
			url: 'api/v1/notify/test',
			method: 'POST',
			body: JSON.stringify({
				type: getValues(`${path}.type`),
				...deepDiff(dataJSON, original),
				...extras,
				service_id: getValues('id'),
				service_name: getValues('name'),
				name_previous: original?.name,
			}),
		});

	const {
		data: testData,
		isFetching,
		refetch: testRefetch,
		status: testStatus,
	} = useQuery({
		queryKey: [
			'test_notify',
			{
				service: extras?.service_id_previous,
				notify: original?.name,
			},
			{
				// ...getValues(path) - shallow copy as convertNotifyToAPI mutates the object.
				params: convertNotifyToAPI({ ...getValues(path) }),
			},
		],
		queryFn: ({ queryKey }) =>
			fetchTestNotifyJSON(
				(queryKey[2] as { params: unknown }).params as NotifyTypesValues,
			),
		enabled: false,
		notifyOnChangeProps: 'all',
		retry: false,
		staleTime: 0,
	});

	const refetch = async () => {
		// Prevent refetching too often.
		const currentTime = Date.now();
		if (currentTime - lastFetched < 2000) return;

		const result = await trigger(path, { shouldFocus: true });
		if (result) {
			testRefetch();
			setLastFetched(currentTime);
		}
	};

	const ResultIcon = useMemo(() => {
		if (isFetching)
			return (
				<FontAwesomeIcon
					icon={faSpinner}
					spin
					style={{ marginLeft: '0.5rem' }}
				/>
			);
		if (testStatus !== 'pending') {
			const err = testData?.message !== undefined || testStatus === 'error';
			return (
				<FontAwesomeIcon
					icon={err ? faCircleXmark : faCheckCircle}
					className={`text-${err ? 'danger' : 'success'}`}
				/>
			);
		}
		return <></>;
	}, [isFetching, testStatus, testData]);

	return (
		<span style={{ alignItems: 'center' }}>
			<span className="pt-1 pb-2" style={{ display: 'flex' }}>
				{ResultIcon}
				<Button
					variant="secondary"
					style={{ marginLeft: 'auto', padding: '0 1rem' }}
					onClick={refetch}
					disabled={isFetching}
				>
					<FontAwesomeIcon icon={faSync} style={{ paddingRight: '0.25rem' }} />
					Send Test Message
				</Button>
			</span>
			{/* Render either the server error or form validation error */}
			<Result status={testStatus} err={testData?.message} />
			{errors && (
				<Alert
					variant="danger"
					style={{ paddingLeft: '2rem', marginBottom: 'unset' }}
				>
					{Object.entries(errors).map(([key, error]) => (
						<li key={key}>
							{key}: {error}
						</li>
					))}
				</Alert>
			)}
		</span>
	);
};

export default TestNotify;
