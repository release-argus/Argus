import type { NotifyType } from '@/utils/api/types/config/notify/all-types';
import type { NotifySchemaValues } from '@/utils/api/types/config-edit/notify/schemas';

export type NotifyTestRequestBuilder = {
	/* The service ID. */
	serviceID?: string;
	/* The previous service ID. */
	previousServiceID?: string | null;
	/* The service name. */
	serviceName: string;
	/* The notify type. */
	type: NotifyType;
	/* Extra data to send. */
	extras: Record<string, string>;
	/* The default values. */
	defaults?: NotifySchemaValues;
	/* The previous values. */
	previous?: NotifySchemaValues;
	/* The new values. */
	new: NotifySchemaValues;
};

export type NotifyTestResponse = {
	/* The result of the test. */
	message: string;
};
