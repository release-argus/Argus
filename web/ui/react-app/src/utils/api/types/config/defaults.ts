import type {
	NotifyMap,
	NotifyTypesMap,
} from '@/utils/api/types/config/notify';
import type { ServiceDefault } from '@/utils/api/types/config/service';
import type { WebHook, WebHookMap } from '@/utils/api/types/config/webhook';

export type HardDefaults = {
	service: ServiceDefault;
	notify: NotifyTypesMap;
	webhook: WebHook;
};

export type Defaults = HardDefaults & {
	notify: Partial<NotifyTypesMap>;
};

export type ServiceEditOtherData = {
	defaults: Defaults;
	hard_defaults: Defaults;
	notify: NotifyMap;
	webhook: WebHookMap;
};
