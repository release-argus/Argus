import { MoreHorizontalIcon } from 'lucide-react';
import type { ComponentProps, ReactNode } from 'react';
import { cn } from '@/lib/utils';

function LoadingEllipsis({ className, ...props }: ComponentProps<'span'>) {
	return (
		<span
			aria-hidden
			className={cn('mx-auto flex w-full justify-center', className)}
			{...props}
		>
			<MoreHorizontalIcon className="size-4 animate-pulse" />
			<span className="sr-only">Loading...</span>
		</span>
	);
}

function TextOrLoading({
	loading,
	text,
	loadingElement,
	className,
	...props
}: {
	loading: boolean;
	text: string;
	loadingElement?: ReactNode;
} & ComponentProps<'span'>) {
	return (
		<span
			className={className}
			style={{ display: 'inline-block', minWidth: `${text.length}ch` }}
			{...props}
		>
			{loading ? (loadingElement ?? <LoadingEllipsis />) : text}
		</span>
	);
}

export { LoadingEllipsis, TextOrLoading };
