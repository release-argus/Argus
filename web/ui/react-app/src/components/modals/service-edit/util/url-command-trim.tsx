import { URLCommandType } from "types/config";

// urlCommandTrim will remove any keys not used for the type
export const urlCommandTrim = (command: URLCommandType) => {
  if (command.type === "regex")
    return {
      type: "regex",
      regex: command.regex,
      index: command.index ? Number(command.index) : undefined,
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
}) => {
  return Object.values(commands).map((value) => urlCommandTrim(value));
};
