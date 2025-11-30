/* Query keys for React Query. */
export const QUERY_KEYS = {
	CONFIG: {
		BUILD_INFO: () => ['config/build'],
		CLI_FLAGS: () => ['config/flags'],
		RAW: () => ['config'],
		RUNTIME_INFO: () => ['config/runtime'],
	},
	NOTIFY: {
		TEST: (serviceID?: string | null, notifyID?: string) => [
			'test_notify',
			{
				notify: notifyID,
				service: serviceID,
			},
		],
	},
	SERVICE: {
		ACTIONS: (serviceID: string) => ['service/actions', { service: serviceID }],
		DETAIL: () => ['service/edit', 'detail'],
		EDIT_ITEM: (serviceID: string) => ['service/edit', { service: serviceID }],
		ORDER: () => ['service/order'],
		SUMMARY_ITEM: (serviceID: string) => [
			'service/summary',
			{ service: serviceID },
		],
	},
};
