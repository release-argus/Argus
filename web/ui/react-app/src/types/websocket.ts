import type {
	CommandSummaryListType,
	ServiceSummary,
	WebHookSummaryListType,
} from '@/utils/api/types/config/summary';

export type WebSocketResponse =
	| {
			page: 'APPROVALS';
			type: 'ACTION';
			sub_type: 'SENDING' | 'REFRESH' | 'RESET';
			service_data?: ServiceSummary;
			command_data?: CommandSummaryListType;
			webhook_data?: WebHookSummaryListType;
	  }
	| {
			page: 'APPROVALS';
			type: 'COMMAND' | 'WEBHOOK';
			sub_type: 'EVENT';
			service_data: ServiceSummary;
			command_data?: CommandSummaryListType;
			webhook_data?: WebHookSummaryListType;
	  }
	| {
			page: 'APPROVALS';
			type: 'DELETE';
			sub_type: string;
			order?: string[];
	  }
	| {
			page: 'APPROVALS';
			type: 'EDIT';
			/* The service ID prior to the edit ('' when the service was created). */
			sub_type?: string;
			/* Fields unchanged by the edit are omitted, so `id` is absent unless it
			   changed - resolve the service from `sub_type` when it is. */
			service_data?: Partial<ServiceSummary>;
	  }
	| {
			page: 'APPROVALS';
			type: 'SERVICE';
			sub_type: 'INIT' | 'ORDER';
			order?: string[];
			service_data?: ServiceSummary;
	  }
	| {
			page: 'APPROVALS';
			type: 'VERSION';
			sub_type: 'ACTION' | 'INIT' | 'QUERY' | 'UPDATED' | 'NEW';
			service_data: ServiceSummary;
	  };
