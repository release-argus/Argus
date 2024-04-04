import { URLCommandType } from "types/config";
import { isEmptyOrNull } from "utils";

/**
 * Returns a `url_command` object with only the relevant keys for the type
 *
 * @param command - The URLCommandType to trim
 * @param sending - Whether the command is being sent to the server
 * @returns A URLCommandType with only the relevant keys for the type
 */
export const urlCommandTrim = (
  command: URLCommandType,
  sending: boolean
): URLCommandType => {
  // regex
  if (command.type === "regex")
    if (sending)
      return {
        type: "regex",
        regex: command.regex,
        index: command.index ? Number(command.index) : undefined,
        template: command.template,
      };
    else
      return {
        type: "regex",
        regex: command.regex,
        index: command.index ? Number(command.index) : undefined,
        template: command.template ? command.template : undefined,
        template_toggle: !isEmptyOrNull(command.template),
      };

  // replace
  if (command.type === "replace")
    return { type: "replace", old: command.old, new: command.new };

  // else, it's a split
  return {
    type: "split",
    text: command.text,
    index: command.index ? Number(command.index) : undefined,
  };
};

/**
 * urlCommandsTrim will remove any keys not used for fhe type for all URLCommandTypes in the list
 *
 * @param commands - The URLCommandType[] to trim
 * @returns A URLCommandType[] with only the relevant keys for each type
 */
export const urlCommandsTrim = (commands: {
  [key: string]: URLCommandType;
}): URLCommandType[] => {
  return Object.values(commands).map((value) => urlCommandTrim(value, true));
};

export const urlCommandsTrimArray = (
  commands: URLCommandType[]
): URLCommandType[] => {
  return commands.map((value) => urlCommandTrim(value, false));
};
