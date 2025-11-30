import {
	createContext,
	type ReactNode,
	use,
	useCallback,
	useEffect,
	useMemo,
	useRef,
} from 'react';

type TooltipContextValue = {
	/* Function to register a tooltip close function. */
	register: (closeFn: () => void) => void;
	/* Function to unregister a tooltip close function. */
	unregister: () => void;
};

const TooltipContext = createContext<TooltipContextValue | null>(null);

/* Hook to access the tooltip context. */
export const useTooltipContext = () => {
	const ctx = use(TooltipContext);
	if (!ctx) {
		throw new Error(
			'useTooltipContext must be used within a TooltipProviderGlobal',
		);
	}
	return ctx;
};

type TooltipProviderGlobalProps = {
	/* The content to wrap. */
	children: ReactNode;
};

/**
 * TooltipProviderGlobal provides a global tooltip context.
 *
 * @param children - The content to wrap.
 */
export const TooltipProviderGlobal = ({
	children,
}: TooltipProviderGlobalProps) => {
	const closeFnRef = useRef<(() => void) | null>(null);

	useEffect(() => {
		const handleKeyDown = (e: KeyboardEvent) => {
			if (e.key === 'Escape' && closeFnRef.current) {
				closeFnRef.current();
			}
		};

		document.addEventListener('keydown', handleKeyDown);
		return () => {
			document.removeEventListener('keydown', handleKeyDown);
		};
	}, []);

	const register = useCallback((fn: () => void) => {
		closeFnRef.current = fn;
	}, []);

	const unregister = useCallback(() => {
		closeFnRef.current = null;
	}, []);

	const value = useMemo(
		() => ({
			register,
			unregister,
		}),
		[register, unregister],
	);

	return <TooltipContext value={value}>{children}</TooltipContext>;
};
