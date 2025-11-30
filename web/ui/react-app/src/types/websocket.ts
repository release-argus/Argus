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
			sub_type?: string;
			service_data?: ServiceSummary;
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
