import type { OptionType } from '@/components/ui/react-select/custom-components';

export type OpsGenieAction = {
	arg: string;
}[];

export const OPSGENIE_TARGET_TYPE = {
	TEAM: {
		label: 'Team',
		SUB_TYPE: {
			ID: { label: 'ID', value: 'id' },
			NAME: { label: 'Name', value: 'name' },
		},
		value: 'team',
	},
	USER: {
		label: 'User',
		SUB_TYPE: {
			ID: { label: 'ID', value: 'id' },
			USERNAME: { label: 'Username', value: 'username' },
		},
		value: 'user',
	},
} as const;
// "team" | "user"
type OpsGenieTargetType =
	(typeof OPSGENIE_TARGET_TYPE)[keyof typeof OPSGENIE_TARGET_TYPE]['value'];
export const OpsGenieTargetTypeOptions = Object.values(
	OPSGENIE_TARGET_TYPE,
).map((t) => ({ label: t.label, value: t.value }));
// "team": "id" | "name" -- "user": "id" | "username"
type SubTypeValues<T> = T extends {
	SUB_TYPE: Record<string, { value: infer V }>;
}
	? V
	: never;
export const OpsGenieTargetSubTypeOptions = Object.fromEntries(
	Object.values(OPSGENIE_TARGET_TYPE).map((target) => [
		target.value,
		Object.values(target.SUB_TYPE),
	]),
) as Record<OpsGenieTargetType, OptionType[]>;
export type OpsGenieTargetSubType<T extends keyof typeof OPSGENIE_TARGET_TYPE> =
	SubTypeValues<(typeof OPSGENIE_TARGET_TYPE)[T]>;

// Team
type OpsGenieTargetTeam = {
	sub_type: OpsGenieTargetSubType<'TEAM'>;
	type: typeof OPSGENIE_TARGET_TYPE.TEAM.value;
	value: string;
};
// User
type OpsGenieTargetUser = {
	sub_type: OpsGenieTargetSubType<'USER'>;
	type: typeof OPSGENIE_TARGET_TYPE.USER.value;
	value: string;
};
export type OpsGenieTarget = OpsGenieTargetTeam | OpsGenieTargetUser;
export type OpsGenieTargets = OpsGenieTarget[];
