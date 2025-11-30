// biome-ignore assist/source/useSortedKeys: none on top.
export const TELEGRAM_PARSEMODE = {
	NONE: { label: 'None', value: 'None' },
	HTML: { label: 'HTML', value: 'HTML' },
	MARKDOWN: { label: 'Markdown', value: 'Markdown' },
	MARKDOWN_V2: { label: 'Markdown v2', value: 'MarkdownV2' },
} as const;
export type TelegramParsemode =
	(typeof TELEGRAM_PARSEMODE)[keyof typeof TELEGRAM_PARSEMODE]['value'];
export const telegramParsemodeOptions = Object.values(TELEGRAM_PARSEMODE);
