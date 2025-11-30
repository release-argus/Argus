// SOURCE: https://gist.github.com/ilkou/7bf2dbd42a7faf70053b43034fc4b5a4

import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import {
	CaretSortIcon,
	CheckIcon,
	Cross2Icon as CloseIcon,
} from '@radix-ui/react-icons';
import type {
	ClearIndicatorProps,
	DropdownIndicatorProps,
	GroupBase,
	MultiValueProps,
	MultiValueRemoveProps,
	OptionProps,
} from 'react-select';
import { components } from 'react-select';
import type { ReadonlyKeys } from '@/lib/types';

export type OptionTriStatus = 'include' | 'exclude' | undefined;
export type OptionType = {
	label: string;
	value: string;
	status?: string;
};
export type OptionReadonly = ReadonlyKeys<
	OptionType,
	'label' | 'value' | 'status'
>;

export type TriGetStatus = (v: string) => OptionTriStatus;
const getTriGetStatus = (selectProps: unknown): TriGetStatus | undefined => {
	if (
		selectProps &&
		typeof selectProps === 'object' &&
		'triGetStatus' in (selectProps as Record<string, unknown>)
	) {
		const fn = (selectProps as { triGetStatus?: unknown }).triGetStatus;
		return typeof fn === 'function' ? (fn as TriGetStatus) : undefined;
	}
	return undefined;
};

export const DropdownIndicator = <
	Option,
	IsMulti extends boolean,
	Group extends GroupBase<Option>,
>(
	props: DropdownIndicatorProps<Option, IsMulti, Group>,
) => {
	return (
		<components.DropdownIndicator {...props}>
			<CaretSortIcon className={'h-4 w-4 opacity-50'} />
		</components.DropdownIndicator>
	);
};

export const ClearIndicator = <
	Option,
	IsMulti extends boolean,
	Group extends GroupBase<Option>,
>(
	props: ClearIndicatorProps<Option, IsMulti, Group>,
) => {
	return (
		<components.ClearIndicator
			{...props}
			innerProps={{
				...props.innerProps,
				style: {
					...(props.innerProps?.style ?? {}),
					cursor: 'pointer',
				},
			}}
		>
			<CloseIcon className={'h-3.5 w-3.5 opacity-50'} />
		</components.ClearIndicator>
	);
};

export const MultiValueRemove = <
	Option,
	IsMulti extends boolean,
	Group extends GroupBase<Option>,
>(
	props: MultiValueRemoveProps<Option, IsMulti, Group>,
) => {
	return (
		<components.MultiValueRemove
			{...props}
			innerProps={{
				...props.innerProps,
				style: {
					...(props.innerProps?.style ?? {}),
					cursor: 'pointer',
				},
			}}
		>
			<CloseIcon className={'h-3 w-3 opacity-50'} />
		</components.MultiValueRemove>
	);
};

export const MultiValueTriState = <
	Option extends OptionType,
	IsMulti extends boolean,
	Group extends GroupBase<Option>,
>(
	props: MultiValueProps<Option, IsMulti, Group>,
) => {
	const triGetStatus = getTriGetStatus(props.selectProps);
	const status = triGetStatus
		? triGetStatus(props.data.value)
		: props.data.status; // fallback

	return (
		<components.MultiValue {...props}>
			<div className="flex items-center gap-1">
				{status === 'include' && <CheckIcon className="size-4 shrink-0" />}
				{status === 'exclude' && <CloseIcon className="size-4 shrink-0" />}
				<span>{props.data.label}</span>
			</div>
		</components.MultiValue>
	);
};

export const Option = <
	Option,
	IsMulti extends boolean,
	Group extends GroupBase<Option>,
>(
	props: OptionProps<Option, IsMulti, Group>,
) => {
	return (
		<components.Option {...props}>
			<div className="flex items-center justify-between">
				<div>{props.label}</div>
				{props.isSelected && <CheckIcon className="shrink-0" />}
			</div>
		</components.Option>
	);
};

// Tri-state option renderer used by SelectTriState. Displays an icon based on option.data.status.
export const OptionTriState = <
	Option extends OptionType,
	IsMulti extends boolean,
	Group extends GroupBase<Option>,
>(
	props: OptionProps<Option, IsMulti, Group>,
) => {
	const triGetStatus = getTriGetStatus(props.selectProps);
	const status = triGetStatus
		? triGetStatus(props.data.value)
		: props.data.status; // fallback

	return (
		<components.Option {...props}>
			<div className="flex items-center justify-between">
				<div>{props.label}</div>
				{status === 'include' && <CheckIcon className="shrink-0" />}
				{status === 'exclude' && <CloseIcon className="shrink-0" />}
			</div>
		</components.Option>
	);
};

// Wraps the react-select MultiValue to make it draggable.
export const SortableMultiValue = <
	FinalOptionType extends OptionType,
	IsMulti extends boolean,
	Group extends GroupBase<FinalOptionType>,
>(
	props: MultiValueProps<FinalOptionType, IsMulti, Group>,
) => {
	const {
		attributes,
		listeners,
		setNodeRef,
		transform,
		transition,
		isDragging,
	} = useSortable({ id: props.data.value });

	// Style the component during drag and drop operations.
	const style = {
		cursor: 'move',
		opacity: isDragging ? 0.5 : 1,
		transform: CSS.Transform.toString(transform),
		transition: transition,
	};

	// Pass the sortable props to the innerProps of the MultiValue component.
	return (
		<components.MultiValue
			{...props}
			innerProps={{
				...props.innerProps,
				ref: setNodeRef,
				style,
				...attributes,
				...listeners,
			}}
		>
			{props.children}
		</components.MultiValue>
	);
};
