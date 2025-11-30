import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import {
	smtpAuthOptions,
	smtpEncryptionOptions,
} from '@/utils/api/types/config/notify/smtp';

export const SMTPAuthEnum = z.enum(toZodEnumTuple(smtpAuthOptions));

export const SMTPEncryptionEnum = z.enum(toZodEnumTuple(smtpEncryptionOptions));
