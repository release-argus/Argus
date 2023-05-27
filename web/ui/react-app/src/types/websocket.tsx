import {
  CommandSummaryListType,
  ServiceSummaryType,
  WebHookSummaryListType,
} from "./summary";

export interface GorillaWebSocketMessage {
  bubbles: boolean;
  cancelBubble: boolean;
  cancelable: boolean;
  composed: boolean;
  currentTarget: null;
  data: WebSocketResponse;
  defaultPrevented: boolean;
  eventPhase: number;
  explicitOriginalTarget: ExplicitOriginalTarget;
  isTrusted: boolean;
  lastEventId: string;
  origin: string;
  originalTarget: OriginalTarget;
  ports: number[];
  returnValue: boolean;
  source: null;
  srcElement: SRCElement;
  target: Target;
  timeStamp: number;
  type: string;
}

export interface WebSocketResponse {
  page: string;
  type: string;
  sub_type?: string;
  order?: string[];
  target?: string;
  service_data?: ServiceSummaryType;
  command_data?: CommandSummaryListType;
  webhook_data?: WebHookSummaryListType;
}

export interface ExplicitOriginalTarget {
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

export interface OriginalTarget {
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

export interface SRCElement {
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

export interface Target {
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
