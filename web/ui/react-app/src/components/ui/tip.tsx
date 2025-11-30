import {
	type ComponentProps,
	isValidElement,
	type PropsWithChildren,
	type ReactNode,
	useEffect,
	useRef,
	useState,
} from 'react';
import { Button } from '@/components/ui/button';
import { Toggle } from '@/components/ui/toggle';
import {
	Tooltip,
	TooltipContent,
	TooltipProvider,
	TooltipTrigger,
} from '@/components/ui/tooltip';
import { useTooltipContext } from '@/hooks/use-tooltip';
import { cn } from '@/lib/utils';

const touchDelay = (lastTouch: number) => {
	return lastTouch && Date.now() - lastTouch < 25;
};

const Tip = ({
	content,
	children,
	className,
	contentProps,
	delayDuration = 0,
	touchDelayDuration = 0,
}: PropsWithChildren<{
	content: string | ReactNode;
	className?: string;
	contentProps?: Omit<ComponentProps<typeof TooltipContent>, 'className'>;
	delayDuration?: number;
	touchDelayDuration?: number;
}>) => {
	const { register, unregister } = useTooltipContext();
	const [open, setOpen] = useState(false);
	const timeoutRef = useRef<NodeJS.Timeout | null>(null);
	const [lastTouch, setLastTouch] = useState(0);

	const openTooltip = () => {
		setOpen(true);
	};
	const closeTooltip = () => {
		setOpen(false);
	};
	const toggleTooltip = () => {
		setOpen((prev) => !prev);
	};

	const handleClick = () => {
		if (touchDelay(lastTouch)) return;

		toggleTooltip();
	};

	const handleMouseEnter = () => {
		if (touchDelay(lastTouch)) return;
		if (timeoutRef.current) clearTimeout(timeoutRef.current);

		if (delayDuration) {
			timeoutRef.current = setTimeout(openTooltip, delayDuration);
		} else {
			openTooltip();
		}
	};

	const handleMouseLeave = () => {
		if (timeoutRef.current) clearTimeout(timeoutRef.current);

		timeoutRef.current = setTimeout(closeTooltip, 500);
	};

	const handleFocus = () => {
		if (touchDelay(lastTouch)) return;

		openTooltip();
	};

	const handleBlur = () => {
		if (touchDelay(lastTouch)) return;
		if (timeoutRef.current) {
			clearTimeout(timeoutRef.current);
		}

		timeoutRef.current = setTimeout(closeTooltip, 250);
	};

	const handleTouchStart = () => {
		setLastTouch(Date.now());

		if (touchDelayDuration) {
			timeoutRef.current = setTimeout(toggleTooltip, touchDelayDuration);
		} else {
			toggleTooltip();
		}
	};

	const handleTouchEnd = () => {
		setLastTouch(Date.now());
		if (!timeoutRef.current) return;

		clearTimeout(timeoutRef.current);
		timeoutRef.current = null;
	};

	// biome-ignore lint/correctness/useExhaustiveDependencies: register stable.
	useEffect(() => {
		if (!open) return undefined;
		register(() => {
			setOpen(false);
		});
		return () => {
			unregister();
		};
	}, [open]);

	const isButton =
		isValidElement(children) &&
		(children.type === Button || children.type === Toggle);

	return (
		<TooltipProvider>
			<Tooltip delayDuration={delayDuration} open={open}>
				<TooltipTrigger asChild>
					{isButton ? (
						<span
							className={cn('inline-flex', className)}
							onBlur={handleBlur}
							onFocus={handleFocus}
							onMouseEnter={handleMouseEnter}
							onMouseLeave={handleMouseLeave}
							onTouchEnd={handleTouchEnd}
							onTouchMove={handleTouchEnd}
							onTouchStart={handleTouchStart}
						>
							{children}
						</span>
					) : (
						<Button
							className={cn(
								'cursor-default hover:bg-transparent! hover:text-inherit hover:opacity-80',
								className,
							)}
							onBlur={handleBlur}
							onClick={handleClick}
							onFocus={handleFocus}
							onMouseEnter={handleMouseEnter}
							onMouseLeave={handleMouseLeave}
							onTouchEnd={handleTouchEnd}
							onTouchMove={handleTouchEnd}
							onTouchStart={handleTouchStart}
							size="fit"
							type="button"
							variant="ghost"
						>
							{children}
						</Button>
					)}
				</TooltipTrigger>
				<TooltipContent
					className={cn(!content && 'hidden', 'max-w-[60vw] text-wrap p-2')}
					{...(contentProps ?? {})}
				>
					<span className="inline-block">{content}</span>
				</TooltipContent>
			</Tooltip>
		</TooltipProvider>
	);
};

export default Tip;
