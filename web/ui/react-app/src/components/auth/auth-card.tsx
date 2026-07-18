import type { ReactNode } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import getBasename from '@/utils/get-basename';

type AuthCardProps = {
	/** Shown beside the title. */
	icon: ReactNode;
	title: string;
	/** Sub-heading below the title. */
	description?: string;
	/** The card body (the auth form). */
	children: ReactNode;
};

/** The centred logo + card layout shared by the login and first-run-setup pages. */
export const AuthCard = ({
	icon,
	title,
	description,
	children,
}: AuthCardProps) => (
	<div className="flex min-h-svh w-full flex-col items-center justify-center gap-4 p-4 sm:gap-6 sm:p-5">
		<div className="flex select-none items-center gap-3">
			<img
				alt="Argus"
				className="size-10 sm:size-14"
				src={`${getBasename()}/favicon.svg`}
			/>
			<span
				aria-hidden
				className="font-semibold text-3xl text-brand sm:text-4xl"
			>
				Argus
			</span>
		</div>
		<Card className="w-full max-w-sm">
			<CardHeader>
				<CardTitle className="flex items-center gap-2 text-2xl">
					{icon}
					<h1>{title}</h1>
				</CardTitle>
				{description && (
					<p className="text-muted-foreground text-sm">{description}</p>
				)}
			</CardHeader>
			<CardContent className="[&_input]:scroll-mb-[30dvh]">
				{children}
			</CardContent>
		</Card>
	</div>
);
