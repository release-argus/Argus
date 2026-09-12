import { z } from 'zod';
import {
	approvalsToolbarViewOptions,
	type CardTimestampType,
	type HideName,
	type ToolbarViewOption,
	toolbarHideOptions,
	toolbarTimestampOptions,
} from '@/constants/toolbar';

const HIDE_NAMES = toolbarHideOptions.map(({ key }) => key) as [
	HideName,
	...HideName[],
];
const TIMESTAMP_NAMES = toolbarTimestampOptions.map(({ key }) => key) as [
	CardTimestampType,
	...CardTimestampType[],
];
const VIEW_VALUES = approvalsToolbarViewOptions.map(({ value }) => value) as [
	ToolbarViewOption,
	...ToolbarViewOption[],
];

export const defaultsSchema = z.object({
	hide: z.array(z.enum(HIDE_NAMES)),
	timestamps: z.array(z.enum(TIMESTAMP_NAMES)),
	view: z.enum(VIEW_VALUES),
});

export type DefaultsFormValues = z.infer<typeof defaultsSchema>;
