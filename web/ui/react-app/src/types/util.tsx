export interface Dictionary<T> {
  [Key: string]: T;
}

export interface ApprovalsToolbarOptions {
  [key: string]: string | boolean | number[];

  search: string;
  editMode: boolean;

  // 0 - hide up-to-date.
  // 1 - hide updatable.
  // 2 - hide skipped.
  // 3 - hide inactive.
  hide: number[];
}

export interface ReactChangeEvent {
  target: {
    name: string;
    value?: string | boolean | number | string[];
  };
}

export interface OptionType {
  label: string;
  value: string;
}
