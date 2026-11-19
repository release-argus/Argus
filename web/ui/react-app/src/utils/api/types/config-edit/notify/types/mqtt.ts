import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import { mqttQoSOptions } from '@/utils/api/types/config/notify/mqtt';

export const MQTTQoSZodEnum = z.enum(toZodEnumTuple(mqttQoSOptions));
