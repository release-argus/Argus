import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import { signalgridTypeOptions } from '@/utils/api/types/config/notify/signalgrid';

export const SignalgridTypeZodEnum = z.enum(
	toZodEnumTuple(signalgridTypeOptions),
);
