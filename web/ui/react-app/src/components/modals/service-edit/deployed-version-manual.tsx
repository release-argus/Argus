import { useMutation } from '@tanstack/react-query';
import { LoaderCircle, Save } from 'lucide-react';
import { memo, useState } from 'react';
import { Controller, useFormContext, useWatch } from 'react-hook-form';
import { toast } from 'sonner';
import { FieldLabel } from '@/components/generic/field';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Field, FieldGroup } from '@/components/ui/field';
import {
	InputGroup,
	InputGroupAddon,
	InputGroupButton,
	InputGroupInput,
} from '@/components/ui/input-group';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import { useWebSocket } from '@/contexts/websocket';
import useValuesRefetch from '@/hooks/values-refetch';
import { cn } from '@/lib/utils';
import { beautifyGoErrors } from '@/utils';
import { mapRequest } from '@/utils/api/types/api-request-handler';

/* The throttle time for saving the version. */
const SAVE_THROTTLE_MS = 1000;

/**
 * The `deployed_version` form fields for 'manual' version.
 */
const DeployedVersionManual = () => {
	const name = 'deployed_version.version';
	const { serviceID, schemaData } = useSchemaContext();
	const { monitorData } = useWebSocket();
	const { control } = useFormContext();

	const [lastFetched, setLastFetched] = useState(0);

	const versionField = useWatch({ name: name }) as string;
	const { data: semanticVersioning, refetchData: refetchSemanticVersioning } =
		useValuesRefetch<boolean>('options.semantic_versioning');

	const original = schemaData?.deployed_version;
	const originalOptions = schemaData?.options;
	const service = serviceID ? monitorData.service[serviceID] : undefined;
	const status = service?.status ?? {};

	const canSave =
		serviceID &&
		original?.type === 'manual' &&
		versionField !== status.deployed_version;

	const handleSave = async () => {
		// Prevent refetching too often.
		const currentTime = Date.now();
		if (currentTime - lastFetched < SAVE_THROTTLE_MS || !versionField) return;

		refetchSemanticVersioning();

		// setTimeout to allow time for the refetch states above.
		await new Promise((resolve) => setTimeout(resolve, 0));

		setLastFetched(currentTime);
		try {
			const data = await saveVersion();
			status.deployed_version = data.version;
		} catch (error) {
			console.error('Failed to save version', error);
			toast.error('Failed to save version:', {
				description: mutationError?.message,
			});
		}
	};

	const {
		mutateAsync: saveVersion,
		error: mutationError,
		isPending: isSaving,
	} = useMutation({
		mutationFn: () =>
			mapRequest('VERSION_REFRESH', {
				data: { type: 'manual', version: versionField },
				dataSemanticVersioning: semanticVersioning ?? null,
				dataTarget: 'deployed_version',
				original: original,
				originalSemanticVersioning:
					originalOptions?.semantic_versioning ?? null,
				serviceID: serviceID,
			}),
	});

	return (
		<>
			<FieldGroup className="col-span-6 py-1 lg:col-span-10">
				<Controller
					control={control}
					name={name}
					render={({ field, fieldState }) => (
						<Field data-invalid={fieldState.invalid}>
							<FieldLabel
								htmlFor={name}
								text="Version"
								tooltip={{
									content: 'The version that you have deployed',
									type: 'string',
								}}
							/>
							<InputGroup>
								<InputGroupInput
									{...field}
									className={cn(
										mutationError &&
											'border-destructive focus:ring-destructive',
									)}
									id="version"
									type="text"
								/>
								{canSave && field.value && (
									<InputGroupAddon align="inline-end" className="!py-0 !pr-1.5">
										<InputGroupButton
											aria-label="Save version"
											className="rounded-l-none border-l"
											disabled={isSaving || !field.value}
											onClick={handleSave}
											size="icon-md"
											variant="outline"
										>
											{isSaving ? (
												<LoaderCircle className="h-full animate-spin" />
											) : (
												<Save />
											)}
										</InputGroupButton>
									</InputGroupAddon>
								)}
							</InputGroup>
						</Field>
					)}
				/>
			</FieldGroup>
			{mutationError && (
				<span className="col-span-full mb-2 w-full break-all pt-1">
					<Alert variant="destructive">
						<AlertTitle>Failed to save version:</AlertTitle>
						<AlertDescription>
							{beautifyGoErrors(mutationError.message)}
						</AlertDescription>
					</Alert>
				</span>
			)}
		</>
	);
};

export default memo(DeployedVersionManual);
