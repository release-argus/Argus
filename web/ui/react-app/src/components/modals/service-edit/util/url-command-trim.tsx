import { URLCommandType } from "types/config";

// urlCommandTrim will remove any keys not used for the type
export const urlCommandTrim = (command: URLCommandType, sending: boolean) => {
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
        template_toggle: (command.template || "") !== "",
      };
  if (command.type === "replace")
    return { type: "replace", old: command.old, new: command.new };
  // else, it's a split
  return {
    type: "split",
    text: command.text,
    index: command.index ? Number(command.index) : undefined,
  };
};

// urlCommandsTrim will remove any unsued keye for the type for all URLCommandTypes in the list
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
