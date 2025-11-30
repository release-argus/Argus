import { type FunctionComponent, type ReactElement, useMemo } from 'react';
import SelectComponent, {
	type ActionMeta,
	type GroupBase,
	type OnChangeValue,
} from 'react-select';
import CreatableSelectComponent from 'react-select/creatable';
import {
	ClearIndicator,
	DropdownIndicator,
	MultiValueRemove,
	MultiValueTriState,
	OptionTriState,
	type OptionTriStatus,
	type OptionType,
	type TriGetStatus,
} from '@/components/ui/react-select/custom-components';
import {
	getSelectDetails,
	onBlurWorkaround,
} from '@/components/ui/react-select/helper';
import type { SelectProps } from '@/components/ui/react-select/select';

type ClickState = { value: string; state: NonNullable<OptionTriStatus> };
/**
 * Determines whether a given value is an array of `ClickState` objects.
 *
 * A valid `ClickState` object must have:
 * - A `value` property of type `string`.
 * - A `state` property that is either 'include' or 'exclude'.
 *
 * @param v - The value to check for conformity to the `ClickState[]` type.
 * @returns `true` if the input is an array of `ClickState` objects, otherwise `false`.
 */
const isClickStateList = (v: unknown): v is ClickState[] => {
	// Not a list.
	if (!Array.isArray(v)) return false;
	// Empty list.
	if (v.length === 0) return true;

	// Test the first element.
	const first = v[0];
	if (typeof first !== 'object' || first === null) return false;

	const maybeValue = (first as { value?: unknown }).value;
	const maybeState = (first as { state?: unknown }).state;
	return (
		typeof maybeValue === 'string' &&
		(maybeState === 'include' || maybeState === 'exclude')
	);
};

/**
 * The next state after the current state.
 */
const nextState: Record<
	NonNullable<OptionTriStatus>,
	NonNullable<OptionTriStatus> | 'remove'
> = {
	exclude: 'remove',
	include: 'exclude',
};

type TriSelectProps<
	Option extends OptionType,
	IsMulti extends boolean = false,
	Group extends GroupBase<Option> = GroupBase<Option>,
> = Omit<SelectProps<Option, IsMulti, Group>, 'value' | 'onChange'> & {
	value?: ClickState[];
	onChange?: (list: ClickState[]) => void;
};

/**
 * A Tr-State Select component for rendering a dropdown input.
 * It supports single-select and multi-select modes and optional 'creatable' functionality.
 *
 * @template OptionType - The shape of an individual option in the dropdown.
 * @template IsMulti - Boolean indicating if the select should permit multiple values.
 * @template Group - Represents the shape of the group definition for grouped options.
 *
 * @param props - Props for configuring the Select component.
 * @param ref - Object reference.
 * @param isCreatable - Determines if users can add new options to the dropdown.
 * @param fixedHeight - Fix the height of the component.
 * @param props.options - The list of options for the select component.
 * @param props.value - The selected value or values.
 * @param props.onChange - Callback function triggered whenever the value changes.
 * @param props.components - Custom component overrides for various select elements.
 * @param props.styles - Custom styling options for the component.
 * @param props.classNames - Custom className overrides for internal elements of the select component.
 * @param props.rest - Additional props to extend the component functionality as needed.
 *
 * @return A renderable Select component.
 */
const SelectTriState = <
	Option extends OptionType,
	IsMulti extends boolean = false,
	Group extends GroupBase<Option> = GroupBase<Option>,
>({
	ref,
	isCreatable = false,
	fixedHeight = false,
	...props
}: TriSelectProps<Option, IsMulti, Group>): ReactElement => {
	// React-Select styling.
	// biome-ignore lint/correctness/useExhaustiveDependencies: fixedHeight, props.className, and props.styles stable.
	const details = useMemo(
		() =>
			getSelectDetails<Option, IsMulti, Group>({
				classNames: props.classNames,
				fixedHeight,
				styles: props.styles,
			}),
		[],
	);

	const {
		value,
		onChange,
		options = [],
		components = {},
		isMulti = false,
		...rest
	} = props;

	const clickList = isClickStateList(value) ? value : [];

	// Cycle the option at the given value.
	// - remove: remove the option from the list no matter its current state.
	const cycle = (val: string, remove = false) => {
		const idx = clickList.findIndex((c) => c.value === val);
		let nextList: ClickState[];

		if (idx === -1) {
			// Add this new option.
			nextList = [...clickList, { state: 'include', value: val }];
		} else {
			// New state after cycle/remove.
			const newState = remove ? 'remove' : nextState[clickList[idx].state];
			nextList =
				newState === 'remove'
					? clickList.filter((_, i) => i !== idx)
					: clickList.map((c, i) =>
							i === idx ? { ...c, state: newState } : c,
						);
		}

		if (onChange) onChange(nextList);
	};

	const Component = (
		isCreatable
			? CreatableSelectComponent
			: SelectComponent<Option, IsMulti, Group>
	) as FunctionComponent<
		SelectProps<Option, IsMulti, Group> & { triGetStatus?: TriGetStatus }
	>;

	const clickMap = new Map(clickList.map((c) => [c.value, c.state] as const));
	const selectedValue = (
		isMulti
			? (options as Option[]).filter((o) => clickMap.has(o.value))
			: ((options as Option[]).find((o) => clickMap.has(o.value)) ?? null)
	) as IsMulti extends true ? readonly Option[] : Option | null;

	return (
		<Component
			classNames={details.classNames}
			components={{
				ClearIndicator,
				DropdownIndicator,
				MultiValue: MultiValueTriState,
				MultiValueRemove,
				Option: OptionTriState,
				...components,
			}}
			isMulti={isMulti as IsMulti}
			onChange={(
				_nv: OnChangeValue<Option, IsMulti>,
				meta: ActionMeta<Option>,
			) => {
				if (
					meta.action === 'select-option' ||
					meta.action === 'deselect-option'
				) {
					if (meta.option) cycle(meta.option.value);
				} else if (meta.action === 'remove-value') {
					cycle(meta.removedValue.value, true);
				} else if (meta.action === 'clear') {
					if (onChange) onChange([]);
				}
			}}
			options={options}
			ref={ref}
			styles={details.styles}
			triGetStatus={(v: string) => clickList.find((c) => c.value === v)?.state}
			unstyled
			value={selectedValue}
			{...rest}
			closeMenuOnSelect={false}
			hideSelectedOptions={false}
			onBlur={onBlurWorkaround}
		/>
	);
};
SelectTriState.displayName = 'Select-Tri-State';

export default SelectTriState as <
	FinalOptionType extends OptionType = OptionType,
	IsMulti extends boolean = false,
	Group extends GroupBase<FinalOptionType> = GroupBase<FinalOptionType>,
>(
	p: TriSelectProps<FinalOptionType, IsMulti, Group>,
) => ReactElement;
