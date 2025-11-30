import { useQuery } from '@tanstack/react-query';
import { AlertCircleIcon, LoaderCircle } from 'lucide-react';
import { type FC, use, useCallback, useEffect, useMemo, useState } from 'react';
import { type FieldErrors, FormProvider, useWatch } from 'react-hook-form';
import type z from 'zod';
import { HelpTooltip } from '@/components/generic';
import EditService from '@/components/modals/service-edit/service';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from '@/components/ui/dialog';
import { TextOrLoading } from '@/components/ui/loading-ellipsis';
import { ModalContext } from '@/contexts/modal';
import {
	SchemaProvider,
	useSchemaContext,
} from '@/contexts/service-edit-zod-type';
import useServiceForm from '@/hooks/use-service-form';
import { useServiceEdit } from '@/hooks/use-service-mutation';
import { QUERY_KEYS } from '@/lib/query-keys';
import { DeleteModal } from '@/modals/delete-confirm';
import { extractErrors } from '@/utils';
import { mapRequest } from '@/utils/api/types/api-request-handler';
import { mapServiceToAPIRequest } from '@/utils/api/types/config-edit/service/api/conversions';
import { getErrorMessage } from '@/utils/errors';

/**
 * @returns The service edit modal.
 */
const ServiceEditModal = () => {
	const { modal } = use(ModalContext);
	if (modal.actionType !== 'EDIT') {
		return null;
	}
	return <ServiceEditModalGetData serviceID={modal.service.id} />;
};

type ServiceEditModalGetDataProps = {
	/* The ID of the service to edit. */
	serviceID: string;
};
/**
 * Gets the data for and returns the service edit modal
 *
 * @param serviceID - The ID of the service to edit.
 * @returns The service edit modal, showing a loading state whilst fetching.
 */
const ServiceEditModalGetData: FC<ServiceEditModalGetDataProps> = ({
	serviceID,
}) => {
	const [loadingModal, setLoadingModal] = useState(true);
	useEffect(() => {
		const timeout = setTimeout(() => {
			setLoadingModal(false);
		}, 200);
		return () => {
			clearTimeout(timeout);
		};
	}, []);

	// Fetch the defaults/hardDefaults and notify/webhook globals.
	const { data: otherOptionsData, isFetched: isFetchedOtherOptionsData } =
		useQuery({
			queryFn: () => mapRequest('SERVICE_EDIT_DEFAULTS', null),
			queryKey: QUERY_KEYS.SERVICE.DETAIL(),
		});
	// Fetch the existing service data.
	const { data: serviceData, isSuccess: isSuccessServiceData } = useQuery({
		enabled: !!serviceID,
		queryFn: () => mapRequest('SERVICE_EDIT_DETAIL', { serviceID: serviceID }),
		queryKey: QUERY_KEYS.SERVICE.EDIT_ITEM(serviceID),
	});

	const hasFetched =
		isFetchedOtherOptionsData &&
		(isSuccessServiceData || !serviceID) &&
		otherOptionsData !== undefined;

	return (
		<SchemaProvider data={serviceData} otherOptionsData={otherOptionsData}>
			<ServiceEditModalWithData
				loading={loadingModal || !hasFetched}
				serviceID={serviceID}
			/>
		</SchemaProvider>
	);
};

type ServiceEditModalHeaderProps = {
	type: 'Create' | 'Edit';
};

/**
 * @returns The header for the service edit modal.
 */
const ServiceEditModalHeader: FC<ServiceEditModalHeaderProps> = ({ type }) => (
	<DialogHeader className="flex items-center justify-between">
		<DialogTitle>
			<strong>{`${type} Service`}</strong>
			<HelpTooltip
				content="Greyed out placeholder text represents a default that you can override.
					(current secrets can be kept by leaving them as '<secret>')."
				delayDuration={500}
				type="string"
			/>
		</DialogTitle>
		<DialogDescription className="sr-only">
			Configure the service options.
		</DialogDescription>
	</DialogHeader>
);

type ServiceEditModalWithDataProps = {
	/* Whether the modal is fetching data. */
	loading: boolean;
	/* The ID of the service to edit. */
	serviceID?: string;
};

/**
 * A modal for editing a service.
 *
 * @param serviceID - The ID of the service to edit.
 * @param loading - Indicates whether the modal shows a loading state.
 * @returns A modal for editing a service.
 */
const ServiceEditModalWithData: FC<ServiceEditModalWithDataProps> = ({
	serviceID,
	loading,
}) => {
	const { setModal } = use(ModalContext);

	// biome-ignore lint/correctness/useExhaustiveDependencies: setModal stable
	const hideModal = useCallback(() => {
		setModal({ actionType: '', service: { id: '', loading: true } });
	}, []);
	const {
		schema,
		schemaData,
		schemaDataDefaults,
		mainDataDefaults,
		serviceID: sID,
	} = useSchemaContext();

	const form = useServiceForm({
		defaultValues: {
			comment: '',
			id: '',
			name: '',
			...schemaData,
		},
		schema: schema,
	});

	// biome-ignore lint/correctness/useExhaustiveDependencies: form stable. Reset only once, after schemaData loaded.
	useEffect(() => {
		if (sID === undefined) return;
		if (schemaData) form.reset(schemaData);
	}, [sID, loading]);
	// null if submitting.
	const [err, setErr] = useState<string | null>('');
	const onError = useCallback((error: unknown) => {
		setErr(getErrorMessage(error));
	}, []);

	const { mutateAsync, isPending: isSubmitting } = useServiceEdit(schema, {
		onError: onError,
		onSuccess: hideModal,
	});
	// biome-ignore lint/correctness/useExhaustiveDependencies: form stable
	const onClick = useCallback(() => {
		void form.handleSubmit(async (data: z.infer<typeof schema>) => {
			const dataParsed = schema.safeParse(data);
			if (!dataParsed.success) {
				console.error(dataParsed.error);
				return;
			}
			const dataPayload = mapServiceToAPIRequest(
				dataParsed.data,
				schemaDataDefaults,
			);

			try {
				await mutateAsync({ data: dataPayload, serviceID: serviceID ?? null });
			} catch (error) {
				const message = getErrorMessage(error);
				console.error(error);
				form.setError('root', { message, type: 'server' });
			}
		})();
	}, [mainDataDefaults, mutateAsync, schema, schemaDataDefaults, serviceID]);

	const separateNameToggle = useWatch({
		control: form.control,
		name: 'id_name_separator',
	});

	// Format the errors.
	const errors = useMemo(
		() => renameErrorField(form.formState.errors, separateNameToggle),
		[separateNameToggle, form.formState.errors],
	);

	return (
		<Dialog key={serviceID} onOpenChange={hideModal} open={true}>
			<FormProvider {...form}>
				<form>
					<DialogContent className="max-h-full w-full max-w-full overflow-y-auto sm:max-w-xl md:max-h-[95%] md:max-w-2xl lg:max-w-4xl">
						<ServiceEditModalHeader type={serviceID ? 'Edit' : 'Create'} />
						<EditService loading={loading} />
						<DialogFooter className="flex flex-col">
							<div className="flex w-full items-center justify-between">
								<div>
									{serviceID && (
										<DeleteModal disabled={err === null || loading} />
									)}
								</div>
								{err === null && <LoaderCircle className="animate-spin" />}
								<div className="flex">
									<Button
										disabled={err === null || loading}
										id="modal-cancel"
										onClick={() => hideModal()}
										variant="secondary"
									>
										Cancel
									</Button>
									<Button
										className="ms-2"
										disabled={
											err === null ||
											!form.formState.isDirty ||
											loading ||
											isSubmitting
										}
										id="modal-action"
										onClick={onClick}
										type="submit"
									>
										<TextOrLoading loading={isSubmitting} text="Confirm" />
									</Button>
								</div>
							</div>
							{form.formState.submitCount > 0 &&
								(!form.formState.isValid || err) && (
									<Alert className="mb-0 pl-8" variant="destructive">
										<AlertCircleIcon />
										<AlertTitle>
											Please correct the errors in the form and try again.
										</AlertTitle>
										<AlertDescription>
											{/* Render either the server error or form validation error */}
											{err ? (
												<>
													{err.split(`\\n`).map((line) => (
														<pre
															className="whitespace-pre-wrap break-words"
															key={line}
														>
															{line}
														</pre>
													))}
												</>
											) : (
												<ul className="list-inside list-disc">
													{Object.entries(extractErrors(errors) ?? []).map(
														([key, error]) => (
															<li key={key}>
																{key}: {error}
															</li>
														),
													)}
												</ul>
											)}
										</AlertDescription>
									</Alert>
								)}
						</DialogFooter>
					</DialogContent>
				</form>
			</FormProvider>
		</Dialog>
	);
};

export default ServiceEditModal;

/**
 * Renames error fields in the form state according to whether `id` and `name` fields remain separate.
 * When `separateNameField` equals `false`, converts `id` errors to `name` errors.
 * Ensures `name` appears first in the error list (after `id`, when present).
 *
 * @param errors - The form field errors.
 * @param separateNameField - Controls separation of the `id` and `name` fields.
 * @returns The error object with fields renamed as needed.
 */
const renameErrorField = (
	errors: FieldErrors,
	separateNameField: boolean,
): FieldErrors => {
	if (!('id' in errors || 'name' in errors)) return errors;

	const entries: [string, unknown][] = [];
	// Rename 'id' to 'name' when we have no id/name separation.
	if ('id' in errors)
		entries.push([separateNameField ? 'id' : 'name', errors.id]);
	// Push 'name' first when fields remain separate.
	if (separateNameField && 'name' in errors)
		entries.push(['name', errors.name]);

	for (const [key, value] of Object.entries(errors)) {
		if (key === 'id' || key === 'name') continue;

		entries.push([key, value]);
	}

	return Object.fromEntries(entries) as FieldErrors;
};
