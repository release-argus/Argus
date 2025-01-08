import {
	CommandSummaryListType,
	ServiceSummaryType,
	WebHookSummaryListType,
} from './summary';

export interface GorillaWebSocketMessage {
	bubbles: boolean;
	cancelBubble: boolean;
	cancelable: boolean;
	composed: boolean;
	currentTarget: null;
	data: WebSocketResponse;
	defaultPrevented: boolean;
	eventPhase: number;
	explicitOriginalTarget: WebSocketMessageDetail;
	isTrusted: boolean;
	lastEventId: string;
	origin: string;
	originalTarget: WebSocketMessageDetail;
	ports: number[];
	returnValue: boolean;
	source: null;
	srcElement: WebSocketMessageDetail;
	target: WebSocketMessageDetail;
	timeStamp: number;
	type: string;
}

export type WebSocketResponse =
	| {
			page: 'APPROVALS';
			type: 'ACTION';
			sub_type: 'SENDING' | 'REFRESH' | 'RESET';
			service_data?: ServiceSummaryType;
			command_data?: CommandSummaryListType;
			webhook_data?: WebHookSummaryListType;
	  }
	| {
			page: 'APPROVALS';
			type: 'COMMAND' | 'WEBHOOK';
			sub_type: 'EVENT';
			service_data: ServiceSummaryType;
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
			sub_type: string;
			service_data?: ServiceSummaryType;
	  }
	| {
			page: 'APPROVALS';
			type: 'SERVICE';
			sub_type: 'INIT' | 'ORDER';
			order?: string[];
			service_data?: ServiceSummaryType;
	  }
	| {
			page: 'APPROVALS';
			type: 'VERSION';
			sub_type: 'ACTION' | 'INIT' | 'QUERY' | 'UPDATED' | 'NEW';
			service_data: ServiceSummaryType;
	  };

export interface WebSocketMessageDetail {
	binaryType: string;
	bufferedAmount: number;
	extensions: string;
	onclose: null;
	onerror: null;
	onmessage: null;
	onopen: null;
	protocol: string;
	readyState: number;
	url: string;
}
