import { type FC, memo } from 'react';
import { REGEX, REPLACE, SPLIT } from '.';
import type { URLCommand } from '@/utils/api/types/config/service/latest-version';

const RENDER_TYPE_COMPONENTS: Record<
	URLCommand['type'],
	FC<{
		name: string;
	}>
> = {
	regex: REGEX,
	replace: REPLACE,
	split: SPLIT,
};

/**
 * Form fields for this 'type' of `url_command`.
 *
 * @param name - The name of the `url_command` in the form.
 * @param commandType - Specifies the `url_command` type.
 */
const RenderURLCommand = ({
	name,
	commandType,
}: {
	name: string;
	commandType: URLCommand['type'];
}) => {
	const RenderTypeComponent = RENDER_TYPE_COMPONENTS[commandType];
	if (!RenderTypeComponent) return null;

	return <RenderTypeComponent name={name} />;
};

export default memo(RenderURLCommand);
