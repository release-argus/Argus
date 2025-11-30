/* Make all keys in T readonly. */
export type ReadonlyKeys<T, K extends keyof T> = {
	readonly [P in K]: T[P];
} & {
	[P in Exclude<keyof T, K>]: T[P];
};
