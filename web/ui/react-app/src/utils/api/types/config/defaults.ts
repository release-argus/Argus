import type {
	NotifyMap,
	NotifyTypesMap,
} from '@/utils/api/types/config/notify';
import type { ServiceDefault } from '@/utils/api/types/config/service';
import type { Webhook, WebhookMap } from '@/utils/api/types/config/webhook';

export type HardDefaults = {
	service: ServiceDefault;
	notify: NotifyTypesMap;
	webhook: Webhook;
};

export type Defaults = HardDefaults & {
	notify: Partial<NotifyTypesMap>;
};

export type ServiceEditOtherData = {
	defaults: Defaults;
	hard_defaults: Defaults;
	notify: NotifyMap;
	webhook: WebhookMap;
};
