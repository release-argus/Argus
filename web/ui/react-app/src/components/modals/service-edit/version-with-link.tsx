import { Link } from 'lucide-react';
import { type FC, useMemo } from 'react';
import type { ColSize } from '@/components/generic/field-shared';
import FieldTextWithButton from '@/components/generic/field-text-with-button';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import { getServiceURL } from '@/components/modals/service-edit/util';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import {
	LATEST_VERSION_LOOKUP_TYPE,
	type LatestVersionLookupType,
} from '@/utils/api/types/config/service/latest-version';
import type { DeployedVersionURLSchema } from '@/utils/api/types/config-edit/service/types/deployed-version';
import type { LatestVersionLookupSchema } from '@/utils/api/types/config-edit/service/types/latest-version';
import {
	type NullString,
	nullString,
} from '@/utils/api/types/config-edit/shared/null-string';
import { isValidForgeHost } from '@/utils/api/types/config-edit/validators';

const typeConfig = {
	forgejo: {
		buttonAriaLabel: 'Open repository',
		isURL: false,
		label: 'Repository',
	},
	github: {
		buttonAriaLabel: 'Open repository',
		isURL: false,
		label: 'Repository',
	},
	url: {
		buttonAriaLabel: 'Open URL',
		isURL: true,
		label: 'URL',
	},
};

type BaseProps = {
	/* The name of the field in the form. */
	name: string;
	/* The 'type' of the version field. */
	type: LatestVersionLookupType | NullString;
	/* Whether the field is required. */
	required?: boolean;

	/* The column width on different screen sizes. */
	colSize?: ColSize;
};

type VersionWithLinkProps = BaseProps & {
	/* The tooltip on the field label. */
	tooltip: TooltipWithAriaProps;
	/* The instance the repository is hosted on (Forgejo only). */
	host?: string;
};

/**
 * The 'version field', with a link to the source being monitored.
 *
 * @param name - The name of the field in the form.
 * @param type - The 'type' of version field.
 * @param required - Whether the field is required.
 * @param host - The instance the repository is hosted on (Forgejo only).
 * @param tooltip - The tooltip on the field label.
 * @param tooltip.type - 'string' | 'element'.
 * @param tooltip.side - The wide to render the tooltip content.
 * @param tooltip.size - The size of the tooltip.
 * @param tooltip.delayDuration - Time before rendering the tooltip.
 *
 * @param colSize - The column width on different screen sizes.
 *
 * @returns The version field with a link to the source being monitored.
 */
const VersionWithLink: FC<VersionWithLinkProps> = ({
	name,
	type,
	required,
	host,
	tooltip,
	colSize,
}) => {
	const { schemaDataDefaults } = useSchemaContext();

	// biome-ignore lint/correctness/useExhaustiveDependencies: schemaDataDefaults stable.
	const resolvedType = useMemo(() => {
		const key = name.split('.')[0] as keyof typeof schemaDataDefaults;
		const keyDefaults =
			schemaDataDefaults?.[key as keyof typeof schemaDataDefaults];

		let typedKeyDefaults: LatestVersionLookupSchema | DeployedVersionURLSchema;
		if (key === 'latest_version') {
			typedKeyDefaults = keyDefaults as LatestVersionLookupSchema;
		} else {
			typedKeyDefaults = keyDefaults as DeployedVersionURLSchema;
		}

		return (
			(type === nullString ? typedKeyDefaults.type : type) ??
			LATEST_VERSION_LOOKUP_TYPE.GITHUB.value
		);
	}, [name, type]);
	const config = typeConfig[resolvedType];

	// A forge repository is only addressable once the instance is known.
	const hasDestination =
		resolvedType !== LATEST_VERSION_LOOKUP_TYPE.FORGEJO.value ||
		isValidForgeHost(host);

	return (
		<FieldTextWithButton
			button={
				hasDestination
					? {
							ariaLabel: config.buttonAriaLabel,
							href: (value: string) => getServiceURL(resolvedType, value, host),
							Icon: Link,
							kind: 'link',
						}
					: undefined
			}
			colSize={colSize}
			label={config.label}
			name={name}
			required={required}
			tooltip={tooltip}
			type="text"
		/>
	);
};

export default VersionWithLink;
