/* Query keys for React Query. */
export const QUERY_KEYS = {
	CONFIG: {
		BUILD_INFO: () => ['config', 'build'],
		CLI_FLAGS: () => ['config', 'flags'],
		RAW: () => ['config'],
		RUNTIME_INFO: () => ['config', 'runtime'],
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
		ACTIONS: (serviceID: string) => [
			...QUERY_KEYS.SERVICE.BASE(serviceID),
			'actions',
		],
		BASE: (serviceID: string) => ['service', { service: serviceID }],
		EDIT_DEFAULTS: () => ['service', 'edit', 'defaults'],
		EDIT_ITEM: (serviceID: string) => [
			...QUERY_KEYS.SERVICE.BASE(serviceID),
			'edit',
		],
		ORDER: () => ['service', 'order'],
		SUMMARY_ITEM: (serviceID: string) => [
			...QUERY_KEYS.SERVICE.SUMMARY_ITEM_BASE,
			{ service: serviceID },
		],
		SUMMARY_ITEM_BASE: ['service', 'summary'],
	},
};
