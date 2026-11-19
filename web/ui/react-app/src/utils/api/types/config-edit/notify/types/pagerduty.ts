import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import {
	pagerDutyActionOptions,
	pagerDutySeverityOptions,
} from '@/utils/api/types/config/notify/pagerduty';

export const PagerDutyActionZodEnum = z.enum(
	toZodEnumTuple(pagerDutyActionOptions),
);

export const PagerDutySeverityZodEnum = z.enum(
	toZodEnumTuple(pagerDutySeverityOptions),
);
