export type Command = string[];
export type Commands = Command[];

export type CustomHeader = {
	old_index: number | null;
	key: string;
	value: string;
};
export type CustomHeaders = CustomHeader[];

export type EmptyObject = Record<string, never>;
