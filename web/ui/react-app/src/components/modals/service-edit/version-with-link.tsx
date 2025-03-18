import { FC } from 'react';
import FormTextWithButton from 'components/generic/form-text-with-button';
import { Position } from 'types/config';
import { TooltipWithAriaProps } from 'components/generic/tooltip';
import { faLink } from '@fortawesome/free-solid-svg-icons';
import { repoTest } from 'components/generic/form-validate';

interface Props {
	name: string;
	type: 'github' | 'url';
	required?: boolean;

	col_xs?: number;
	col_sm?: number;
	col_md?: number;
	col_lg?: number;

	positionXS?: Position;
	positionSM?: Position;
	positionMD?: Position;
	positionLG?: Position;
}

type VersionWithLinkProps = TooltipWithAriaProps & Props;

/**
 * The version field with a link to the source being monitored.
 *
 * @param name - The name of the field in the form.
 * @param type - The type of version field.
 * @param required - Whether the field is required.
 *
 * @param col_xs - The amount of columns the item takes up on XS+ screens.
 * @param col_sm - The amount of columns the item takes up on SM+ screens.
 * @param col_md - The amount of columns the item takes up on MD+ screens.
 * @param col_lg - The amount of columns the item takes up on LG+ screens.
 *
 * @param tooltip - The tooltip for the field.
 *
 * @param positionXS - The position of the field on XS+ screens.
 * @param positionSM - The position of the field on SM+ screens.
 * @param positionMD - The position of the field on MD+ screens.
 * @param positionLG - The position of the field on LG+ screens.
 * @returns The version field with a link to the source being monitored.
 */
const VersionWithLink: FC<VersionWithLinkProps> = ({
	name,
	type,
	required,
	col_xs = 12,
	col_sm = 6,
	positionXS = 'left',
	...props
}) => {
	const typeConfig = {
		github: {
			label: 'Repository',
			getLink: (value: string) => `https://github.com/${value}`,
			buttonAriaLabel: 'Open GitHub repository',
			validationFunc: (value: string) => repoTest(value, true),
			isURL: false,
		},
		url: {
			label: 'URL',
			getLink: (value: string) => value,
			buttonAriaLabel: 'Open URL',
			validationFunc: undefined,
			isURL: true,
		},
	};

	const config = typeConfig[type];

	return (
		<FormTextWithButton
			name={name}
			required={required}
			col_xs={col_xs}
			col_sm={col_sm}
			label={config.label}
			type="text"
			isURL={config.isURL}
			validationFunc={config.validationFunc}
			positionXS={positionXS}
			buttonIcon={faLink}
			buttonAriaLabel={config.buttonAriaLabel}
			buttonHref={config.getLink}
			{...props}
		/>
	);
};

export default VersionWithLink;
