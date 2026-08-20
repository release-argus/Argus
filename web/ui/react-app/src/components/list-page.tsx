import { Plus } from 'lucide-react';
import type { ReactElement, ReactNode } from 'react';
import { ListError } from '@/components/list-error';
import { Button } from '@/components/ui/button';
import { Table, TableBody, TableHeader, TableRow } from '@/components/ui/table';

type ListPageProps = {
	/** Page heading. */
	title: string;
	/** Add/create button label. */
	addLabel: string;
	/** Opens the create dialog. */
	onAdd: () => void;
	/** Optional blurb rendered under the header. */
	description?: ReactNode;
	/** Accessible name for the table. */
	tableLabel: string;
	/** Resource name for the error message (lower-case plural). */
	errorResource: string;
	/** Whether the list query failed. */
	isError: boolean;
	/** The list query's error (shown when isError). */
	error: unknown;
	/** The table's [TableHead] cells. */
	columns: ReactNode;
	/** The table body's rows. */
	children: ReactNode;
};

/**
 * Cell padding and header weight.
 */
const TABLE_STYLE = 'border [&_td]:px-4 [&_th]:px-4 [&_th]:font-bold';

/** Shared scaffold for the resource list pages (Users, Groups, API Tokens). */
export const ListPage = ({
	title,
	addLabel,
	onAdd,
	description,
	tableLabel,
	errorResource,
	isError,
	error,
	columns,
	children,
}: ListPageProps): ReactElement => (
	<>
		<div className="flex items-center justify-between pb-2">
			<h2 className="scroll-m-20 font-semibold text-3xl tracking-tight">
				{title}
			</h2>
			<Button aria-label={addLabel} onClick={onAdd}>
				<Plus aria-hidden /> {addLabel}
			</Button>
		</div>
		{description}
		{isError ? (
			<ListError error={error} resource={errorResource} />
		) : (
			<Table aria-label={tableLabel} className={TABLE_STYLE}>
				<TableHeader>
					<TableRow>{columns}</TableRow>
				</TableHeader>
				<TableBody>{children}</TableBody>
			</Table>
		)}
	</>
);
