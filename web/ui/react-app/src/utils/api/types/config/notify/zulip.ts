export const ZULIP_TYPE = {
	CHANNEL: { label: 'Channel', value: 'channel' },
	DIRECT: { label: 'Direct', value: 'direct' },
} as const;
export type ZulipType = (typeof ZULIP_TYPE)[keyof typeof ZULIP_TYPE]['value'];
export const zulipTypeOptions = Object.values(ZULIP_TYPE);
