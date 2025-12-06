import { createContext, type FC, type ReactNode, use, useMemo } from 'react';
import { useServiceOrder } from '@/hooks/use-service-order.ts';
import { useServices } from '@/hooks/use-services.ts';
import type { ServiceEditOtherData } from '@/utils/api/types/config/defaults';
import type { Service } from '@/utils/api/types/config/service';
import { buildServiceSchemaWithFallbacks } from '@/utils/api/types/config-edit/service/form/builder';

type SharedProps = {
	/* The service data to be edited. */
	data?: Service;
	/* Other options data (defaults/hardDefaults and notify/webhook globals) */
	otherOptionsData?: ServiceEditOtherData;
};

type State = SharedProps & {
	/* The service ID. */
	serviceID?: string | null;
	/* The schema for this service. */
	schema: ReturnType<typeof buildServiceSchemaWithFallbacks>['schema'];
	/* The initial form data for this service. */
	schemaData: ReturnType<typeof buildServiceSchemaWithFallbacks>['schemaData'];
	/* The fallback form data. */
	schemaDataDefaults: ReturnType<
		typeof buildServiceSchemaWithFallbacks
	>['schemaDataDefaults'];
	/* Hollow version of the fallback form data. */
	schemaDataDefaultsHollow: ReturnType<
		typeof buildServiceSchemaWithFallbacks
	>['schemaDataDefaultsHollow'];
	/* Type-specific notify/webhook form data. */
	typeDataDefaults: ReturnType<
		typeof buildServiceSchemaWithFallbacks
	>['typeDataDefaults'];
	/* Hollow version of the type-specific notify/webhook form data. */
	typeDataDefaultsHollow: ReturnType<
		typeof buildServiceSchemaWithFallbacks
	>['typeDataDefaultsHollow'];
	/* The notify/webhook globals */
	mainDataDefaults: ReturnType<
		typeof buildServiceSchemaWithFallbacks
	>['mainDataDefaults'];
};

const SchemaContext = createContext<State | null>(null);

type SchemaProviderProps = SharedProps & {
	/* The content to wrap. */
	children: ReactNode;
};

/**
 * SchemaProvider provides the form schema and data for a service create/edit.
 *
 * @param data - The service data to edit.
 * @param otherOptionsData - Other options data (defaults/hardDefaults and notify/webhook globals)
 * @param children - The content to wrap.
 */
export const SchemaProvider: FC<SchemaProviderProps> = ({
	data,
	otherOptionsData,
	children,
}) => {
	const { data: orderData } = useServiceOrder();
	const order = orderData?.order ?? [];
	const services = useServices();

	// Stable key for service IDs.
	// biome-ignore lint/correctness/useExhaustiveDependencies: orderData cocers order.
	const serviceIDs = useMemo(() => {
		const ids = [...order]; // Shallow copy.
		ids.sort((a, b) => a.localeCompare(b, undefined, { sensitivity: 'base' }));
		return ids.join('|');
	}, [orderData]);

	// biome-ignore lint/correctness/useExhaustiveDependencies: serviceIDs covers orderData.order.
	const contextValue = useMemo(() => {
		const serviceNamesSet = new Set(
			services.map((svc) => svc.data?.name).filter(Boolean) as string[],
		);

		const {
			schema,
			schemaData,
			schemaDataDefaults,
			schemaDataDefaultsHollow,
			typeDataDefaults,
			typeDataDefaultsHollow,
			mainDataDefaults,
		} = buildServiceSchemaWithFallbacks(
			serviceNamesSet,
			order,
			otherOptionsData,
			data,
		);

		return {
			data,
			mainDataDefaults,
			otherOptionsData,
			schema,
			schemaData,
			schemaDataDefaults,
			schemaDataDefaultsHollow,
			serviceID:
				schemaData?.id === undefined ? undefined : schemaData.id || null,
			typeDataDefaults,
			typeDataDefaultsHollow,
		};
	}, [data, otherOptionsData, serviceIDs]);

	return <SchemaContext value={contextValue}>{children}</SchemaContext>;
};

/**
 * useSchemaContext retrieves the current schema context value for the form from the SchemaContext.
 */
export const useSchemaContext = () => {
	const context = use(SchemaContext);
	if (!context)
		throw new Error('useSchemaContext must be inside SchemaProvider');

	return context;
};
