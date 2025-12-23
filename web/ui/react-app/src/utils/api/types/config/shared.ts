export type Command = string[];
export type Commands = Command[];

export type Header = {
	old_index: number | null;
	key: string;
	value: string;
};
export type Headers = Header[];

export type EmptyObject = Record<string, never>;
