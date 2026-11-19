// biome-ignore assist/source/useSortedKeys: ascending severity.
export const SIGNALGRID_TYPE = {
	SUCCESS: { label: 'Success', value: 'SUCCESS' },
	INFO: { label: 'Info', value: 'INFO' },
	WARN: { label: 'Warn', value: 'WARN' },
	CRIT: { label: 'Critical', value: 'CRIT' },
} as const;
export type SignalgridType =
	(typeof SIGNALGRID_TYPE)[keyof typeof SIGNALGRID_TYPE]['value'];
export const signalgridTypeOptions = Object.values(SIGNALGRID_TYPE);
