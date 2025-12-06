import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import { telegramParsemodeOptions } from '@/utils/api/types/config/notify/telegram';

export const TelegramParseModeEnum = z.enum(
	toZodEnumTuple(telegramParsemodeOptions),
);
