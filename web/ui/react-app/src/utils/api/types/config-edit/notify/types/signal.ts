import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import { signalTextModeOptions } from '@/utils/api/types/config/notify/signal';

export const SignalTextModeZodEnum = z.enum(
	toZodEnumTuple(signalTextModeOptions),
);
