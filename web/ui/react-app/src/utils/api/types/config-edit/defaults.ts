import type {
	NotifySchemaRecord,
	notifySchemaMap,
} from '@/utils/api/types/config-edit/notify/schemas';
import type { ServiceSchemaDefault } from '@/utils/api/types/config-edit/service/schemas';
import type {
	WebHookSchemaDefault,
	WebHooksSchemaDefault,
} from '@/utils/api/types/config-edit/webhook/schemas';

export type DefaultsSchema = {
	service: ServiceSchemaDefault;
	notify: Partial<typeof notifySchemaMap>;
	webhook: WebHookSchemaDefault;
};

export type ServiceEditOtherDataSchema = {
	defaults: DefaultsSchema;
	hard_defaults: DefaultsSchema;
	notify: NotifySchemaRecord;
	webhook: WebHooksSchemaDefault;
};
