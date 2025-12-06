import { isEmptyOrNull } from '@/utils';
import {
	type FormURLCommand,
	LATEST_VERSION__URL_COMMAND_TYPE,
	type URLCommand,
} from '@/utils/api/types/config/service/latest-version';

/**
 * Trim any unused keys from a `url_command` object.
 *
 * @param command - The URLCommandType to trim.
 * @param apiRequest - For an API request?
 * @param omitValues - Omit values and only set type.
 * @returns A URLCommandType with only the relevant keys for the type.
 */
export function urlCommandTrim(
	command: URLCommand,
	apiRequest: boolean,
	omitValues: true,
): URLCommand;
export function urlCommandTrim(
	command: URLCommand,
	apiRequest: true,
	omitValues: boolean,
): URLCommand;
export function urlCommandTrim(
	command: URLCommand,
	apiRequest: false,
	omitValues: boolean,
): FormURLCommand;
export function urlCommandTrim(
	command: URLCommand,
	apiRequest: boolean,
	omitValues: boolean,
): URLCommand | FormURLCommand {
	switch (command.type) {
		case LATEST_VERSION__URL_COMMAND_TYPE.REGEX.value: {
			let result: URLCommand;
			if (omitValues) {
				result = {
					index: null,
					regex: '',
					template: '',
					type: command.type,
				};
			} else {
				result = {
					index: command.index ?? null,
					regex: command.regex,
					template: command.template ?? '',
					type: command.type,
				};
			}
			if (apiRequest) {
				return result;
			}
			return {
				...result,
				template_toggle: !isEmptyOrNull(command.template),
			} as FormURLCommand;
		}
		case LATEST_VERSION__URL_COMMAND_TYPE.REPLACE.value:
			if (omitValues) {
				return {
					new: '',
					old: '',
					type: command.type,
				};
			}
			return {
				new: command.new,
				old: command.old,
				type: command.type,
			};
		case LATEST_VERSION__URL_COMMAND_TYPE.SPLIT.value: {
			let result: URLCommand;
			if (omitValues) {
				result = {
					index: null,
					text: '',
					type: LATEST_VERSION__URL_COMMAND_TYPE.SPLIT.value,
				};
			} else {
				result = {
					index: command.index ?? null,
					text: command.text,
					type: LATEST_VERSION__URL_COMMAND_TYPE.SPLIT.value,
				};
			}
			if (apiRequest) {
				return result as URLCommand;
			}
			return result as FormURLCommand;
		}
	}
}

/**
 * Trims all unused keys from each URLCommandType.
 *
 * @param commands - The URLCommandType[] to trim.
 * @returns A URLCommandType[] with only the relevant keys for each type.
 */
export const urlCommandsTrim = (
	commands: Record<string, URLCommand>,
): URLCommand[] =>
	Object.values(commands).map((value) => urlCommandTrim(value, true, false));

export const urlCommandsTrimArray = (
	commands: URLCommand[],
	omitValues = false,
): FormURLCommand[] =>
	commands.map((value) => urlCommandTrim(value, false, omitValues));
