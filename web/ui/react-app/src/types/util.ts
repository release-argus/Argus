import type { NullString } from '@/utils/api/types/config-edit/shared/null-string';

/* Screen breakpoints. */
export type ScreenBreakpoint = 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'xxl';

/* Tri-state tags. */
export type TagsTriType = { include: string[]; exclude: string[] };

export type SuccessMessage = {
	message: string;
};

export type StringStringMap = Record<string, string>;
export type StringFieldArray = StringStringMap[];

export const toZodEnumTuple = <T extends readonly { value: string }[]>(
	options: T,
): [T[number]['value'], ...T[number]['value'][]] =>
	options.map((o) => o.value) as [T[number]['value'], ...T[number]['value'][]];

type NonOf<T, U> = T extends U ? never : T;
export type NonNull<T> = NonOf<T, null | undefined | '' | NullString>;
