import type { ReactElement } from 'react';
import { Separator } from '@/components/ui/separator';

type GroupHeadingProps = {
	title: string;
	description: string;
	id?: string;
};

export const GroupHeading = ({
	title,
	description,
	id,
}: GroupHeadingProps): ReactElement => (
	<div className="grid gap-1">
		<h3 className="font-semibold text-base" id={id}>
			{title}
		</h3>
		<p className="text-muted-foreground text-sm">{description}</p>
	</div>
);

export const GroupRule = (): ReactElement => (
	<div className="-mx-4">
		<Separator />
	</div>
);
