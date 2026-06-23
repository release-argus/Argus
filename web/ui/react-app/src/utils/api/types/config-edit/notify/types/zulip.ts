import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import { zulipTypeOptions } from '@/utils/api/types/config/notify/zulip';

export const ZulipTypeZodEnum = z.enum(toZodEnumTuple(zulipTypeOptions));
