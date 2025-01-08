import {
	Button,
	ButtonGroup,
	Container,
	Form,
	Modal,
	Row,
} from 'react-bootstrap';
import {
	FC,
	useCallback,
	useContext,
	useEffect,
	useMemo,
	useState,
} from 'react';
import { FormProvider, useForm } from 'react-hook-form';
import {
	ServiceEditAPIType,
	ServiceEditOtherData,
	ServiceEditType,
} from 'types/service-edit';
import { extractErrors, fetchJSON, removeEmptyValues } from 'utils';

import { DeleteModal } from './delete-confirm';
import EditService from 'components/modals/service-edit/service';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { HelpTooltip } from 'components/generic';
import { ModalContext } from 'contexts/modal';
import { convertAPIServiceDataEditToUI } from 'components/modals/service-edit/util';
import { convertUIServiceDataEditToAPI } from 'components/modals/service-edit/util/ui-api-conversions';
import { faCircleNotch } from '@fortawesome/free-solid-svg-icons';
import { useQuery } from '@tanstack/react-query';

/**
 * @param data - The data to convert.
 * @returns The data with empty values removed and converted to API format.
 */
const getPayload = (data: ServiceEditType) => {
	return removeEmptyValues(convertUIServiceDataEditToAPI(data));
};
/**
 * @returns The service edit modal.
 */
const ServiceEditModal = () => {
	const { handleModal, modal } = useContext(ModalContext);
	if (modal.actionType !== 'EDIT') {
		return null;
	}
	return (
		<ServiceEditModalGetData
			serviceID={modal.service.id}
			hideModal={() => handleModal('', { id: '', loading: true })}
		/>
	);
};

interface ServiceEditModalGetDataProps {
	serviceID: string;
	hideModal: () => void;
}
/**
 * Gets the data for and returns the service edit modal
 *
 * @param serviceID - The ID of the service to edit.
 * @param hideModal - The function to hide the modal.
 * @returns The service edit modal with the data fetched, and editing disabled whilst fetching.
 */
const ServiceEditModalGetData: FC<ServiceEditModalGetDataProps> = ({
	serviceID,
	hideModal,
}) => {
	const [loadingModal, setLoadingModal] = useState(true);
	useEffect(() => {
		const timeout = setTimeout(() => setLoadingModal(false), 200);
		return () => clearTimeout(timeout);
	}, []);
	const { data: otherOptionsData, isFetched: isFetchedOtherOptionsData } =
		useQuery({
			queryKey: ['service/edit', 'detail'],
			queryFn: () =>
				fetchJSON<ServiceEditOtherData>({ url: 'api/v1/service/update' }),
		});
	const { data: serviceData, isSuccess: isSuccessServiceData } = useQuery({
		queryKey: ['service/edit', { id: serviceID }],
		queryFn: () =>
			fetchJSON<ServiceEditAPIType>({
				url: `api/v1/service/update/${encodeURIComponent(serviceID)}`,
			}),
		enabled: !!serviceID,
		refetchOnMount: 'always',
	});

	const hasFetched =
		isFetchedOtherOptionsData &&
		(isSuccessServiceData || !serviceID) &&
		otherOptionsData !== undefined;

	const defaultData: ServiceEditType = useMemo(
		() =>
			convertAPIServiceDataEditToUI(serviceID, serviceData, otherOptionsData),
		[serviceData, otherOptionsData],
	);

	return (
		<ServiceEditModalWithData
			serviceID={serviceID}
			serviceName={serviceData?.name}
			defaultData={defaultData}
			otherOptionsData={otherOptionsData}
			loading={loadingModal || !hasFetched}
			hideModal={hideModal}
		/>
	);
};

/**
 * @returns The header for the service edit modal.
 */
const ServiceEditModalHeader = () => (
	<Modal.Header closeButton>
		<Modal.Title>
			<strong>Edit Service</strong>
			<HelpTooltip
				text="Greyed out placeholder text represents a default that you can override. (current secrets can be kept by leaving them as '<secret>')"
				placement="bottom"
			/>
		</Modal.Title>
	</Modal.Header>
);

interface ServiceEditModalWithDataProps {
	serviceID: string;
	serviceName?: string;
	defaultData: ServiceEditType;
	otherOptionsData?: ServiceEditOtherData;
	loading: boolean;
	hideModal: () => void;
}
/**
 * A modal for editing a service.
 *
 * @param serviceID - The ID of the service to edit.
 * @param serviceName - The name of the service.
 * @param defaultData - The default data for the service.
 * @param otherOptionsData - The mains/defaults/hardDefaults for the service.
 * @param loading - Whether the modal is loading.
 * @param hideModal - The function to hide the modal.
 * @returns A modal for editing a service.
 */
const ServiceEditModalWithData: FC<ServiceEditModalWithDataProps> = ({
	serviceID,
	serviceName,
	defaultData,
	otherOptionsData,
	loading,
	hideModal,
}) => {
	const form = useForm<ServiceEditType>({
		mode: 'onBlur',
		defaultValues: defaultData ?? {},
	});
	useEffect(() => {
		if (defaultData) form.reset(defaultData);
	}, [defaultData]);
	// null if submitting.
	const [err, setErr] = useState<string | null>('');

	const resetAndHideModal = useCallback(() => {
		form.reset({});
		setErr('');
		hideModal();
	}, []);

	const onSubmit = async (data: ServiceEditType) => {
		setErr(null);
		const payload = getPayload(data);

		await fetch(
			serviceID
				? `api/v1/service/update/${encodeURIComponent(serviceID)}`
				: 'api/v1/service/new',
			{
				method: 'PUT',
				body: JSON.stringify(payload),
			},
		)
			.then((response) => {
				if (!response.ok) throw response;
				hideModal();
			})
			.catch(async (err) => {
				let errorMessage = err.statusText;
				try {
					const responseBody = await err.json();
					errorMessage = responseBody.message;
					setErr(errorMessage);
				} catch (e) {
					console.error(e);
					setErr(err.toString());
				}
			});
	};

	const onDelete = async () => {
		console.log(`Deleting ${serviceID}`);
		await fetch(`api/v1/service/delete/${encodeURIComponent(serviceID)}`, {
			method: 'DELETE',
		}).then(() => {
			hideModal();
		});
	};

	return (
		<FormProvider {...form}>
			<Form id="service-edit">
				<Modal size="lg" show animation={false} onHide={resetAndHideModal}>
					<ServiceEditModalHeader />
					<Modal.Body>
						<Container
							fluid
							className="font-weight-bold"
							style={{ paddingLeft: '0.25rem', paddingRight: '0.25rem' }}
						>
							<EditService
								id={serviceID}
								name={serviceName}
								defaultData={defaultData}
								otherOptionsData={otherOptionsData}
								loading={loading}
							/>
						</Container>
					</Modal.Body>
					<Modal.Footer
						style={{ display: 'flex', justifyContent: 'space-between' }}
					>
						<ButtonGroup>
							{serviceID && (
								<DeleteModal
									onDelete={() => onDelete()}
									disabled={err === null || loading}
								/>
							)}
						</ButtonGroup>
						{err === null && (
							<FontAwesomeIcon
								icon={faCircleNotch}
								style={{
									padding: '0',
								}}
								className="fa-spin"
							/>
						)}
						<span>
							<Button
								id="modal-cancel"
								variant="secondary"
								onClick={() => hideModal()}
								disabled={err === null || loading}
							>
								Cancel
							</Button>
							<Button
								id="modal-action"
								variant="primary"
								type="submit"
								onClick={form.handleSubmit(onSubmit)}
								className="ms-2"
								disabled={err === null || !form.formState.isDirty || loading}
							>
								Confirm
							</Button>
						</span>
						{form.formState.submitCount > 0 &&
							(!form.formState.isValid || err) && (
								<Row>
									<div className="error-msg">
										Please correct the errors in the form and try again.
										<br />
										{/* Render either the server error or form validation error */}
										{err ? (
											<>
												{err.split(`\\n`).map((line) => (
													<pre key={line} className="no-margin">
														{line}
													</pre>
												))}
											</>
										) : (
											<ul>
												{Object.entries(
													extractErrors(form.formState.errors) ?? [],
												).map(([key, error]) => (
													<li key={key}>
														{key}: {error}
													</li>
												))}
											</ul>
										)}
									</div>
								</Row>
							)}
					</Modal.Footer>
				</Modal>
			</Form>
		</FormProvider>
	);
};

export default ServiceEditModal;
