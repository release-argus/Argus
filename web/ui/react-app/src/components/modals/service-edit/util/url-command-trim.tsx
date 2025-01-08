import { URLCommandType } from 'types/config';
import { isEmptyOrNull } from 'utils';

/**
 * Trims unused keys from a `url_command` object.
 *
 * @param command - The URLCommandType to trim.
 * @param api_request - For an API request?
 * @returns A URLCommandType with only the relevant keys for the type.
 */
export const urlCommandTrim = (
	command: URLCommandType,
	api_request: boolean,
): URLCommandType => {
	// regex
	if (command.type === 'regex')
		if (api_request)
			return {
				type: 'regex',
				regex: command.regex,
				index: command.index ? Number(command.index) : null,
				template: command.template ?? '',
			};
		else
			return {
				type: 'regex',
				regex: command.regex,
				index: command.index ? Number(command.index) : null,
				template: command.template ?? '',
				template_toggle: !isEmptyOrNull(command.template),
			};

	// replace
	if (command.type === 'replace')
		return { type: 'replace', old: command.old, new: command.new ?? '' };

	// split.
	return {
		type: 'split',
		text: command.text,
		index: command.index ? Number(command.index) : null,
	};
};

/**
 * Trims all unused keys from each URLCommandType.
 *
 * @param commands - The URLCommandType[] to trim.
 * @returns A URLCommandType[] with only the relevant keys for each type.
 */
export const urlCommandsTrim = (commands: {
	[key: string]: URLCommandType;
}): URLCommandType[] => {
	return Object.values(commands).map((value) => urlCommandTrim(value, true));
};

export const urlCommandsTrimArray = (
	commands: URLCommandType[],
): URLCommandType[] => {
	return commands.map((value) => urlCommandTrim(value, false));
};
