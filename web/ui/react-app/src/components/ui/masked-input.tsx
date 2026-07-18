import { Eye, EyeOff } from 'lucide-react';
import { type ComponentProps, useState } from 'react';
import {
	InputGroup,
	InputGroupAddon,
	InputGroupButton,
	InputGroupInput,
} from '@/components/ui/input-group';

type MaskedInputProps = Omit<ComponentProps<typeof InputGroupInput>, 'type'> & {
	/** Noun for the toggle's aria-label, e.g. `password` gives "Show password". */
	valueLabel?: string;
};

/**
 * An input whose value is masked until toggled visible.
 */
export const MaskedInput = ({
	autoComplete = 'new-password',
	valueLabel = 'value',
	...props
}: MaskedInputProps) => {
	const [visible, setVisible] = useState(false);
	return (
		<InputGroup>
			<InputGroupInput
				autoComplete={autoComplete}
				type={visible ? 'text' : 'password'}
				{...props}
			/>
			<InputGroupAddon
				align="inline-end"
				className="h-full py-0 pr-1 has-[>button]:mr-0"
			>
				<InputGroupButton
					aria-label={`${visible ? 'Hide' : 'Show'} ${valueLabel}`}
					className="h-full"
					onClick={() => setVisible((shown) => !shown)}
					size="icon-md"
				>
					{visible ? <EyeOff aria-hidden /> : <Eye aria-hidden />}
				</InputGroupButton>
			</InputGroupAddon>
		</InputGroup>
	);
};
