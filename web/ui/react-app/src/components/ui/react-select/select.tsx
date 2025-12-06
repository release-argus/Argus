// SOURCE: https://gist.github.com/ilkou/7bf2dbd42a7faf70053b43034fc4b5a4

import {
	type ComponentPropsWithoutRef,
	type ComponentRef,
	type ReactElement,
	type RefObject,
	useId,
	useMemo,
} from 'react';
import SelectComponent, { type GroupBase } from 'react-select';
import CreatableSelectComponent from 'react-select/creatable';
import {
	ClearIndicator,
	DropdownIndicator,
	MultiValueRemove,
	Option,
	type OptionType,
} from '@/components/ui/react-select/custom-components';
import {
	getSelectDetails,
	onBlurWorkaround,
} from '@/components/ui/react-select/helper';

type RefType<T> = RefObject<T> | ((instance: T | null) => void);

export type SelectProps<
	Option,
	IsMulti extends boolean,
	Group extends GroupBase<Option> = GroupBase<Option>,
> =
	| ({
			isCreatable: true;
			fixedHeight?: boolean;
	  } & ComponentPropsWithoutRef<
			typeof CreatableSelectComponent<Option, IsMulti, Group>
	  > & {
				ref?: RefType<
					ComponentRef<typeof CreatableSelectComponent<Option, IsMulti, Group>>
				>;
			})
	| ({
			isCreatable?: false;
			fixedHeight?: boolean;
	  } & ComponentPropsWithoutRef<
			typeof SelectComponent<Option, IsMulti, Group>
	  > & {
				ref?: RefType<
					ComponentRef<typeof SelectComponent<Option, IsMulti, Group>>
				>;
			});

/**
 * A Select component for rendering a dropdown input.
 * It supports single-select and multi-select modes, grouped options, and optional creatable functionality.
 *
 * @template OptionType - Type representing the shape of an individual option in the dropdown.
 * @template IsMulti - Boolean type indicating if the select should allow multiple values (default: `false`).
 * @template Group - Type representing the shape of the group definition for grouped options (default: `GroupBase<Option>`).
 *
 * @param props - Props for configuring the Select component.
 * @param ref - Object reference.
 * @param isCreatable - Determines if users can add new options to the dropdown.
 * @param fixedHeight - Fix the height of the component.
 * @param props.options - The list of available options for the select component.
 * @param props.value - The selected value or values.
 * @param props.onChange - Callback function triggered whenever the value changes.
 * @param props.components - Custom component overrides for various select elements, such as indicators and menu.
 * @param props.styles - Custom styling options for the component.
 * @param props.classNames - Custom className overrides for internal elements of the select component.
 * @param props.rest - Additional props to extend the component functionality as needed.
 *
 * @return A renderable Select component.
 */
const Select = <
	Option,
	IsMulti extends boolean = false,
	Group extends GroupBase<Option> = GroupBase<Option>,
>({
	ref,
	isCreatable = false,
	fixedHeight = false,
	...props
}: SelectProps<Option, IsMulti, Group>): ReactElement => {
	// biome-ignore lint/correctness/useExhaustiveDependencies: fixedHeight, props.classNames, and props.styles stable.
	const details = useMemo(
		() =>
			getSelectDetails<Option, IsMulti, Group>({
				classNames: props.classNames,
				fixedHeight,
				styles: props.styles,
			}),
		[],
	);

	const { value, onChange, options = [], components = {}, ...rest } = props;

	const id = useId();
	const Component = isCreatable ? CreatableSelectComponent : SelectComponent;

	return (
		<Component<Option, IsMulti, Group>
			classNames={details.classNames}
			components={{
				ClearIndicator,
				DropdownIndicator,
				MultiValueRemove,
				Option,
				...components,
			}}
			instanceId={id}
			onChange={onChange}
			options={options}
			ref={ref}
			styles={details.styles}
			unstyled
			value={value}
			{...rest}
			onBlur={onBlurWorkaround}
		/>
	);
};
Select.displayName = 'Select';

export default Select as <
	FinalOptionType extends OptionType = OptionType,
	IsMulti extends boolean = false,
>(
	p: SelectProps<FinalOptionType, IsMulti>,
) => ReactElement;
