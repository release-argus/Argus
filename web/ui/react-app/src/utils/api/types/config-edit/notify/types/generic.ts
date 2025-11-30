import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import { genericRequestMethodOptions } from '@/utils/api/types/config/notify/generic';

export const GenericRequestMethodZodEnum = z.enum(
	toZodEnumTuple(genericRequestMethodOptions),
);
