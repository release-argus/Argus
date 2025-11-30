import type { SuccessMessage } from '@/types/util';
import type { ServiceEditOtherData } from '@/utils/api/types/config/defaults';
import type {
	ActionAPIType,
	OrderAPIResponse,
} from '@/utils/api/types/config/summary';
import type {
	ActionGetRequestBuilder,
	ActionSendRequestBuilder,
} from '@/utils/api/types/requests/actions';
import type {
	BuildInfo,
	RuntimeInfo,
} from '@/utils/api/types/requests/app-info';
import type { ConfigGetResponse } from '@/utils/api/types/requests/config';
import type {
	ServiceEditDetailRequestBuilder,
	ServiceEditDetailResponse,
	ServiceSummaryRequestBuilder,
	ServiceSummaryResponse,
} from '@/utils/api/types/requests/defaults';
import type {
	NotifyTestRequestBuilder,
	NotifyTestResponse,
} from '@/utils/api/types/requests/notify-test';
import type {
	ServiceDeleteRequestBuilder,
	ServiceDeleteResponse,
} from '@/utils/api/types/requests/service-delete';
import type {
	ServiceEditRequestBuilder,
	ServiceEditResponse,
} from '@/utils/api/types/requests/service-edit';
import type { ServiceOrderGetRequestBuilder } from '@/utils/api/types/requests/service-order';
import type {
	ServiceRefreshRequest,
	ServiceRefreshRequestBuilder,
	ServiceRefreshResponse,
} from '@/utils/api/types/requests/service-refresh';
import type {
	TemplateParseRequestBuilder,
	TemplateParseResponse,
} from '@/utils/api/types/requests/templates';

type RequestTypeFields<TInput, TRequest, TResponse, TApiResponse = unknown> = {
	/* Params to build the request/response. */
	input: TInput;
	/* Params for the request. */
	request: TRequest;
	/* Final response type. */
	response: TResponse;
	/* Response type from the API. */
	api_response?: TApiResponse;
	/* Response reducer. */
	reducer?: (data: TApiResponse) => TResponse;
};

export type RequestType = {
	ACTION_GET: RequestTypeFields<ActionGetRequestBuilder, null, ActionAPIType>;
	ACTION_SEND: RequestTypeFields<ActionSendRequestBuilder, null, null>;
	CONFIG_GET: RequestTypeFields<null, null, string, ConfigGetResponse>;
	CONFIG_FLAGS: RequestTypeFields<
		null,
		null,
		Record<string, string | boolean | undefined>
	>;
	SERVICE_ORDER_GET: RequestTypeFields<null, null, OrderAPIResponse>;
	SERVICE_ORDER_PUT: RequestTypeFields<
		ServiceOrderGetRequestBuilder,
		null,
		SuccessMessage
	>;
	NOTIFY_TEST: RequestTypeFields<
		NotifyTestRequestBuilder,
		null,
		NotifyTestResponse
	>;
	SERVICE_DELETE: RequestTypeFields<
		ServiceDeleteRequestBuilder,
		null,
		ServiceDeleteResponse
	>;
	SERVICE_EDIT: RequestTypeFields<
		ServiceEditRequestBuilder,
		null,
		ServiceEditResponse
	>;
	SERVICE_EDIT_DEFAULTS: RequestTypeFields<null, null, ServiceEditOtherData>;
	SERVICE_EDIT_DETAIL: RequestTypeFields<
		ServiceEditDetailRequestBuilder,
		null,
		ServiceEditDetailResponse,
		ServiceEditDetailResponse
	>;
	SERVICE_SUMMARY: RequestTypeFields<
		ServiceSummaryRequestBuilder,
		null,
		ServiceSummaryResponse
	>;
	STATUS_BUILD: RequestTypeFields<null, null, BuildInfo>;
	STATUS_RUNTIME: RequestTypeFields<null, null, RuntimeInfo>;
	TEMPLATE_PARSE: RequestTypeFields<
		TemplateParseRequestBuilder,
		null,
		string,
		TemplateParseResponse
	>;
	VERSION_REFRESH: RequestTypeFields<
		ServiceRefreshRequestBuilder,
		ServiceRefreshRequest,
		ServiceRefreshResponse
	>;
};
