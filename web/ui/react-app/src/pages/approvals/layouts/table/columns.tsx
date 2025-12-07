import { ServiceActionRelease } from '@/components/approvals/table/service-action-release';
import { ServiceStatus } from '@/components/approvals/table/service-status';
import type { ColumnDefWithMeta } from '@/components/ui/data-table';
import { DataTableColumnHeader } from '@/components/ui/data-table-column-header';
import { relativeDate } from '@/utils';
import type { ServiceSummary } from '@/utils/api/types/config/summary';
import type { HeaderContext } from '@tanstack/react-table';
import { formatISO9075 } from 'date-fns';

type HeaderContextWithReset<TData, TValue> = HeaderContext<TData, TValue> & {
	resetSorting?: () => void;
};

export const columns: ColumnDefWithMeta<ServiceSummary>[] = [
	{
		accessorKey: 'id',
		enableSorting: true,
		header: (ctx: HeaderContextWithReset<ServiceSummary, unknown>) => (
			<DataTableColumnHeader
				column={ctx.column}
				resetSorting={ctx.resetSorting}
				title="ID"
			/>
		),
		id: 'id',
		meta: { label: 'ID' },
	},
	{
		accessorKey: 'name',
		enableSorting: true,
		header: (ctx: HeaderContextWithReset<ServiceSummary, unknown>) => (
			<DataTableColumnHeader
				column={ctx.column}
				resetSorting={ctx.resetSorting}
				title="Name"
			/>
		),
		id: 'name',
		meta: { hideWhenAllValuesEmpty: true, label: 'Name' },
	},
	{
		accessorFn: (row) =>
			row.has_deployed_version ? row.status?.deployed_version : null,
		cell: ({ row }) =>
			row.original.has_deployed_version
				? row.original.status?.deployed_version
				: null,
		enableSorting: true,
		header: (ctx: HeaderContextWithReset<ServiceSummary, unknown>) => (
			<DataTableColumnHeader
				column={ctx.column}
				resetSorting={ctx.resetSorting}
				title="Deployed Version"
			/>
		),
		id: 'deployed_version',
		meta: { label: 'Deployed Version' },
	},
	{
		accessorFn: (row) =>
			row.has_deployed_version ? row.status?.deployed_version_timestamp : null,
		cell: ({ row }) => (
			<div>
				{row.original.has_deployed_version &&
				row.original.status?.deployed_version_timestamp
					? formatISO9075(
							new Date(row.original.status.deployed_version_timestamp),
						)
					: ''}
			</div>
		),
		enableSorting: true,
		header: (ctx: HeaderContextWithReset<ServiceSummary, unknown>) => (
			<DataTableColumnHeader
				column={ctx.column}
				resetSorting={ctx.resetSorting}
				title="Deployed At"
			/>
		),
		id: 'deployed_version_timestamp',
		meta: { label: 'Deployed At' },
	},
	{
		accessorFn: (row) => row.status?.latest_version ?? null,
		enableSorting: true,
		header: (ctx: HeaderContextWithReset<ServiceSummary, unknown>) => (
			<DataTableColumnHeader
				column={ctx.column}
				resetSorting={ctx.resetSorting}
				title="Latest Version"
			/>
		),
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
		header: (ctx: HeaderContextWithReset<ServiceSummary, unknown>) => (
			<DataTableColumnHeader
				column={ctx.column}
				resetSorting={ctx.resetSorting}
				title="Found At"
			/>
		),
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
		header: (ctx: HeaderContextWithReset<ServiceSummary, unknown>) => (
			<DataTableColumnHeader
				column={ctx.column}
				resetSorting={ctx.resetSorting}
				title="Last Queried"
			/>
		),
		id: 'last_queried',
		meta: { label: 'Last Queried' },
	},
	{
		accessorFn: (row) => row.status?.state ?? null,
		cell: ({ row }) => <ServiceStatus row={row} />,
		enableSorting: true,
		header: (ctx: HeaderContextWithReset<ServiceSummary, unknown>) => (
			<DataTableColumnHeader
				column={ctx.column}
				resetSorting={ctx.resetSorting}
				title="State"
			/>
		),
		id: 'state',
		meta: { label: 'State' },
	},
	{
		cell: ({ row }) => <ServiceActionRelease row={row} />,
		enableSorting: false,
		header: (ctx: HeaderContextWithReset<ServiceSummary, unknown>) => (
			<DataTableColumnHeader
				column={ctx.column}
				resetSorting={ctx.resetSorting}
				title="Actions"
			/>
		),
		id: 'actions',
		meta: { label: 'Actions' },
	},
];
