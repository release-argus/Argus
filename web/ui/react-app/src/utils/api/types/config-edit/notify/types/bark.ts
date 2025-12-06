import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import {
	barkSchemeOptions,
	barkSoundOptions,
} from '@/utils/api/types/config/notify/bark';

export const BarkSchemeEnum = z.enum(toZodEnumTuple(barkSchemeOptions));

export const BarkSoundEnum = z.enum(toZodEnumTuple(barkSoundOptions));
