import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import { barkSoundOptions } from '@/utils/api/types/config/notify/bark';

export const BarkSoundEnum = z.enum(toZodEnumTuple(barkSoundOptions));
