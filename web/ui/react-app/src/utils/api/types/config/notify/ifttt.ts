export const IFTTT_MESSAGE_VALUE = {
	ONE: { label: '1', value: '1' },
	THREE: { label: '3', value: '3' },
	TWO: { label: '2', value: '2' },
} as const;
export type IFTTTMessageValue =
	(typeof IFTTT_MESSAGE_VALUE)[keyof typeof IFTTT_MESSAGE_VALUE]['value'];
export const iftttMessageValueOptions = Object.values(IFTTT_MESSAGE_VALUE);

export const IFTTT_TITLE_VALUE = {
	NONE: { label: 'None', value: '0' },
	ONE: { label: '1', value: '1' },
	THREE: { label: '3', value: '3' },
	TWO: { label: '2', value: '2' },
} as const;
export type IFTTTTitleValue =
	(typeof IFTTT_TITLE_VALUE)[keyof typeof IFTTT_TITLE_VALUE]['value'];
export const iftttTitleValueOptions = Object.values(IFTTT_TITLE_VALUE);
