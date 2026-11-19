// biome-ignore assist/source/useSortedKeys: event lifecycle order.
export const PAGERDUTY_ACTION = {
	TRIGGER: { label: 'Trigger', value: 'trigger' },
	ACKNOWLEDGE: { label: 'Acknowledge', value: 'acknowledge' },
	RESOLVE: { label: 'Resolve', value: 'resolve' },
} as const;
export type PagerDutyAction =
	(typeof PAGERDUTY_ACTION)[keyof typeof PAGERDUTY_ACTION]['value'];
export const pagerDutyActionOptions = Object.values(PAGERDUTY_ACTION);

// biome-ignore assist/source/useSortedKeys: descending severity.
export const PAGERDUTY_SEVERITY = {
	CRITICAL: { label: 'Critical', value: 'critical' },
	ERROR: { label: 'Error', value: 'error' },
	WARNING: { label: 'Warning', value: 'warning' },
	INFO: { label: 'Info', value: 'info' },
} as const;
export type PagerDutySeverity =
	(typeof PAGERDUTY_SEVERITY)[keyof typeof PAGERDUTY_SEVERITY]['value'];
export const pagerDutySeverityOptions = Object.values(PAGERDUTY_SEVERITY);
