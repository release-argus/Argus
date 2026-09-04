import type { ColumnDef, HeaderContext } from '@tanstack/react-table';
import { formatISO9075 } from 'date-fns';
import ServiceImage from '@/components/approvals/service-image';
import { ServiceActionRelease } from '@/components/approvals/table/service-action-release';
import { ServiceID } from '@/components/approvals/table/service-id';
import { ServiceStatus } from '@/components/approvals/table/service-status';
import type { DataTableFeatures } from '@/components/ui/data-table';
import { DataTableColumnHeader } from '@/components/ui/data-table-column-header';
import { relativeDate } from '@/utils';
import type { ServiceSummary } from '@/utils/api/types/config/summary';

/**
 * Builds a column header renderer with sorting and hiding options.
 *
 * @param title - The title of the column.
 */
const columnHeader =
	(title: string) =>
	({ column, table }: HeaderContext<DataTableFeatures, ServiceSummary>) => (
		<DataTableColumnHeader
			column={column}
			resetSorting={table.options.meta?.resetSorting}
			title={title}
		/>
	);

export const columns: ColumnDef<DataTableFeatures, ServiceSummary>[] = [
	{
		accessorKey: 'icon',
		cell: ({ row }) => (
			<ServiceImage className="aspect-square size-8" service={row.original} />
		),
		enableSorting: false,
		header: columnHeader('Icon'),
		id: 'icon',
		meta: { hideWhenAllValuesEmpty: true, label: 'Icon' },
	},
	{
		accessorKey: 'id',
		cell: ({ row }) => <ServiceID row={row} />,
		enableSorting: true,
		header: columnHeader('ID'),
		id: 'id',
		meta: { label: 'ID' },
	},
	{
		accessorKey: 'name',
		enableSorting: true,
		header: columnHeader('Name'),
		id: 'name',
		meta: { hideWhenAllValuesEmpty: true, label: 'Name' },
	},
	{
		accessorFn: (row) =>
			row.deployed_version_type ? row.status?.deployed_version : null,
		cell: ({ row }) =>
			row.original.deployed_version_type
				? row.original.status?.deployed_version
				: null,
		enableSorting: true,
		header: columnHeader('Deployed Version'),
		id: 'deployed_version',
		meta: { label: 'Deployed Version' },
	},
	{
		accessorFn: (row) =>
			row.deployed_version_type ? row.status?.deployed_version_timestamp : null,
		cell: ({ row }) => (
			<div>
				{row.original.deployed_version_type &&
				row.original.status?.deployed_version_timestamp
					? formatISO9075(
							new Date(row.original.status.deployed_version_timestamp),
						)
					: ''}
			</div>
		),
		enableSorting: true,
		header: columnHeader('Deployed At'),
		id: 'deployed_version_timestamp',
		meta: { label: 'Deployed At' },
	},
	{
		accessorFn: (row) => row.status?.latest_version ?? null,
		enableSorting: true,
		header: columnHeader('Latest Version'),
		id: 'latest_version',
		meta: { label: 'Latest Version' },
	},
	{
		accessorFn: (row) => row.status?.latest_version_timestamp ?? null,
		cell: ({ row }) => (
			<div>
				{row.original.status?.latest_version_timestamp
					? formatISO9075(
							new Date(row.original.status.latest_version_timestamp),
						)
					: ''}
			</div>
		),
		enableSorting: true,
		header: columnHeader('Found At'),
		id: 'latest_version_timestamp',
		meta: { label: 'Found At' },
	},
	{
		accessorFn: (row) => row.status?.last_queried ?? null,
		cell: ({ row }) => (
			<div>
				{row.original.status?.last_queried
					? relativeDate(new Date(row.original.status.last_queried))
					: ''}
			</div>
		),

		enableSorting: true,
		header: columnHeader('Last Queried'),
		id: 'last_queried',
		meta: { label: 'Last Queried' },
	},
	{
		accessorFn: (row) => row.status?.state ?? null,
		cell: ({ row }) => <ServiceStatus row={row} />,
		enableSorting: true,
		header: columnHeader('State'),
		id: 'state',
		meta: { label: 'State' },
	},
	{
		cell: ({ row }) => <ServiceActionRelease row={row} />,
		enableSorting: false,
		header: columnHeader('Actions'),
		id: 'actions',
		meta: { label: 'Actions' },
	},
];

/* The IDs of every known column, in the order they are defined above. */
export const COLUMN_IDS: string[] = columns
	.map((col) => col.id ?? ('accessorKey' in col ? String(col.accessorKey) : ''))
	.filter(Boolean);
