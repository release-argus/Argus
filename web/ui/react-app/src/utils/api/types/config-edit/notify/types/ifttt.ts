import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import {
	iftttMessageValueOptions,
	iftttTitleValueOptions,
} from '@/utils/api/types/config/notify/ifttt';

export const IFTTTMessageValueEnum = z.enum(
	toZodEnumTuple(iftttMessageValueOptions),
);
export const IFTTTTitleValueEnum = z.enum(
	toZodEnumTuple(iftttTitleValueOptions),
);
