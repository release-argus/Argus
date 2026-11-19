export const SIGNAL_TEXT_MODE = {
	NONE: { label: 'None', value: 'None' },
	NORMAL: { label: 'Normal', value: 'Normal' },
	STYLED: { label: 'Styled', value: 'Styled' },
} as const;
export type SignalTextMode =
	(typeof SIGNAL_TEXT_MODE)[keyof typeof SIGNAL_TEXT_MODE]['value'];
export const signalTextModeOptions = Object.values(SIGNAL_TEXT_MODE);
