import type { Service } from '@/utils/api/types/config/service';
import type { ServiceSummary } from '@/utils/api/types/config/summary';

export type ServiceEditDetailRequestBuilder = {
	/* The service ID. */
	serviceID: string;
};
export type ServiceEditDetailResponse = Service;

export type ServiceSummaryRequestBuilder = {
	/* The service ID. */
	serviceID: string;
};
export type ServiceSummaryResponse = ServiceSummary;
