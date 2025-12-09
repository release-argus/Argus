import { convertToQueryParams, getChanges } from '@/utils';
import type { RequestType } from '@/utils/api/types/api-request-types';
import type { DeployedVersionLookupManual } from '@/utils/api/types/config/service/deployed-version';
import { mapNotifySchemaToAPIPayload } from '@/utils/api/types/config-edit/notify/api/conversions';
import type { NotifySchemaValues } from '@/utils/api/types/config-edit/notify/schemas';
import type { DeployedVersionLookupSchema } from '@/utils/api/types/config-edit/service/types/deployed-version';
import type { LatestVersionLookupSchema } from '@/utils/api/types/config-edit/service/types/latest-version';
import type {
	ActionGetRequestBuilder,
	ActionSendRequestBuilder,
} from '@/utils/api/types/requests/actions';
import { configReducer } from '@/utils/api/types/requests/config';
import type {
	ServiceEditDetailRequestBuilder,
	ServiceSummaryRequestBuilder,
} from '@/utils/api/types/requests/defaults';
import type { NotifyTestRequestBuilder } from '@/utils/api/types/requests/notify-test';
import type { ServiceDeleteRequestBuilder } from '@/utils/api/types/requests/service-delete';
import {
	type ServiceEditRequestBuilder,
	serviceSummaryReducer,
} from '@/utils/api/types/requests/service-edit';
import type { ServiceOrderGetRequestBuilder } from '@/utils/api/types/requests/service-order';
import type {
	ServiceRefreshRequest,
	ServiceRefreshRequestBuilder,
} from '@/utils/api/types/requests/service-refresh';
import type {
	TemplateParseRequest,
	TemplateParseRequestBuilder,
} from '@/utils/api/types/requests/templates';
import getBasename from '@/utils/get-basename';
import { deepDiff, type GetChangesProps } from '@/utils/query-params';
import removeEmptyValues from '@/utils/remove-empty-values';

export const API_BASE = `${getBasename()}/api/v1`;

type RequestFns = {
	[K in keyof RequestType]: (input: RequestType[K]['input']) => {
		/* HTTP method */
		method: 'DELETE' | 'GET' | 'POST' | 'PUT';
		/* Request headers */
		headers?: Record<string, string>;
		/* Request endpoint */
		endpoint: string;
		/* Request body */
		body?: string;
		/* Request timeout */
		timeout?: number;
		/* Response reducer */
		reducer?: (
			data: RequestType[K]['api_response'],
		) => RequestType[K]['response'];
	};
};

export const RequestMap: RequestFns = {
	ACTION_GET: (input: ActionGetRequestBuilder) => ({
		endpoint: `service/actions/${encodeURIComponent(input.serviceID)}`,
		method: 'GET',
	}),
	ACTION_SEND: (input: ActionSendRequestBuilder) => ({
		body: JSON.stringify({ target: input.target }),
		endpoint: `service/actions/${encodeURIComponent(input.serviceID)}`,
		method: 'POST',
	}),
	CONFIG_FLAGS: () => ({
		endpoint: 'flags',
		method: 'GET',
	}),
	CONFIG_GET: () => ({
		endpoint: 'config',
		method: 'GET',
		reducer: configReducer,
	}),
	NOTIFY_TEST: (input: NotifyTestRequestBuilder) => {
		// Build mapped 'previous' and 'new' using defaults-aware schemas
		const mappedPrevious = input.previous
			? mapNotifySchemaToAPIPayload(input.previous, input.defaults)
			: undefined;
		const mappedNew = mapNotifySchemaToAPIPayload(input.new, input.defaults);

		// Diff the mapped schemas.
		const mappedDiff = mappedPrevious
			? deepDiff(mappedNew, mappedPrevious)
			: mappedNew;

		const notifyPayload = {
			...(mappedDiff as NotifySchemaValues),
			...({ type: input.type } as NotifySchemaValues),
		};

		const payload = {
			...removeEmptyValues({
				name_previous: input.previous?.old_index,
				service_id: input.serviceID,
				service_id_previous: input.previousServiceID,
				service_name: input.serviceName,
				...input.extras,
				type: input.type,
			}),
			...notifyPayload,
		};

		return {
			body: JSON.stringify(payload),
			endpoint: 'notify/test',
			method: 'POST',
		};
	},
	SERVICE_DELETE: (input: ServiceDeleteRequestBuilder) => ({
		endpoint: `service/delete/${encodeURIComponent(input.serviceID)}`,
		method: 'DELETE',
	}),
	SERVICE_EDIT: (input: ServiceEditRequestBuilder) => ({
		body: JSON.stringify(input.body),
		endpoint: input.serviceID
			? `service/update/${encodeURIComponent(input.serviceID)}`
			: `service/new`,
		method: 'PUT',
		timeout: 30000,
	}),
	SERVICE_EDIT_DEFAULTS: () => ({
		endpoint: 'service/update',
		method: 'GET',
	}),
	SERVICE_EDIT_DETAIL: (input: ServiceEditDetailRequestBuilder) => ({
		endpoint: `service/update/${encodeURIComponent(input.serviceID)}`,
		method: 'GET',
		reducer: (response) => ({
			id: input.serviceID,
			...response,
		}),
	}),
	SERVICE_ORDER_GET: () => ({
		endpoint: 'service/order',
		method: 'GET',
	}),
	SERVICE_ORDER_PUT: (input: ServiceOrderGetRequestBuilder) => ({
		body: JSON.stringify({ order: input.order }),
		endpoint: 'service/order',
		method: 'PUT',
	}),
	SERVICE_SUMMARY: (input: ServiceSummaryRequestBuilder) => ({
		endpoint: `service/summary/${encodeURIComponent(input.serviceID)}`,
		method: 'GET',
		reducer: serviceSummaryReducer,
	}),
	STATUS_BUILD: () => ({
		endpoint: 'version',
		method: 'GET',
	}),
	STATUS_RUNTIME: () => ({
		endpoint: 'status/runtime',
		method: 'GET',
	}),
	TEMPLATE_PARSE: (input: TemplateParseRequestBuilder) => ({
		endpoint: `template${convertToQueryParams(
			removeEmptyValues(
				{
					params: input.extraParams
						? JSON.stringify(input.extraParams)
						: undefined,
					service_id: input.serviceID,
					template: input.template,
				} satisfies TemplateParseRequest,
				[],
			) as Record<string, string>,
		)}`,
		method: 'GET',
		reducer: (response) => response?.parsed ?? '',
	}),
	VERSION_REFRESH: (input: ServiceRefreshRequestBuilder) => {
		const props: ServiceRefreshRequest = { queryParams: {} };
		const {
			serviceID,
			data,
			dataSemanticVersioning,
			dataTarget,
			original,
			originalSemanticVersioning,
		} = input;

		const hasSemanticVersioningChanged =
			dataSemanticVersioning !== originalSemanticVersioning;

		if (hasSemanticVersioningChanged) {
			// 'null' signifies default value.
			props.queryParams.semantic_versioning = dataSemanticVersioning ?? 'null';
		}

		if (data) {
			let defaults = original;
			if (
				defaults &&
				Boolean(
					('url' in (defaults as LatestVersionLookupSchema) &&
						!(defaults as LatestVersionLookupSchema).url) ||
						('version' in (defaults as DeployedVersionLookupSchema) &&
							!(defaults as DeployedVersionLookupManual).version),
				)
			) {
				defaults = { ...defaults, type: '' };
			}

			props.queryParams.overrides = getChanges({
				defaults: defaults,
				params: data,
				target: dataTarget,
			} as GetChangesProps);
		}
		const queryParamsStr = convertToQueryParams(props.queryParams);

		const idSegment =
			typeof serviceID === 'string' && serviceID.trim() !== ''
				? `/${encodeURIComponent(serviceID)}`
				: '';

		return {
			endpoint: `${dataTarget}/refresh${idSegment}${queryParamsStr}`,
			method: 'GET',
			timeout: 30000,
		};
	},
};
