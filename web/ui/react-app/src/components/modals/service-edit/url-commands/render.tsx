import { FC, memo } from 'react';
import { REGEX, REPLACE, SPLIT } from '.';

import { URLCommandTypes } from 'types/config';

const RENDER_TYPE_COMPONENTS: {
	[key in URLCommandTypes]: FC<{
		name: string;
	}>;
} = {
	regex: REGEX,
	replace: REPLACE,
	split: SPLIT,
};

/**
 * Renders the form fields for the url_command.
 *
 * @param name - The name of the url_command in the form.
 * @param commandType - The type of the url_command.
 * @returns The form fields for this type of url_command.
 */
const RenderURLCommand = ({
	name,
	commandType,
}: {
	name: string;
	commandType: URLCommandTypes;
}) => {
	const RenderTypeComponent = RENDER_TYPE_COMPONENTS[commandType];

	return <RenderTypeComponent name={name} />;
};

export default memo(RenderURLCommand);
