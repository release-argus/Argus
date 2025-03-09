import { Col, FormControl, FormGroup } from 'react-bootstrap';
import { FC, useCallback, useEffect, useState } from 'react';
import { useFormContext, useWatch } from 'react-hook-form';

import FormLabel from './form-label';
import cx from 'classnames';
import { urlTest } from './form-validate';
import { useError } from 'hooks/errors';

interface Props {
	name: string;

	label: string;
	tooltip?: string;
	isURL?: boolean;
	defaultVal?: string;
	placeholder?: string;
}

/**
 * A form item with a preview image
 *
 * @param name - The name of the form item.
 * @param label - The label of the form item.
 * @param tooltip - The tooltip of the form item.
 * @param isURL - Whether to enable validation for a URL.
 * @param defaultVal - The default value of the form item.
 * @param placeholder - The placeholder of the form item.
 * @returns A form item at name with a preview image, label and tooltip.
 */
const FormTextWithPreview: FC<Props> = ({
	name,

	label,
	tooltip,
	isURL = true,
	defaultVal,
	placeholder,
}) => {
	const { register } = useFormContext();
	const formValue: string | undefined = useWatch({ name: name });
	const error = useError(name, true);

	// The preview image URL, or undefined if invalid.
	const [previewURL, setPreviewURL] = useState(
		urlTest(formValue, true) ? formValue : undefined,
	);
	// Set the preview image.
	const setPreview = useCallback((url?: string) => {
		const previewSource = url || defaultVal;
		if (previewSource && urlTest(previewSource, true) === true) {
			setPreviewURL(previewSource);
		} else {
			setPreviewURL(undefined);
		}
	}, []);

	// Wait for a period of no typing to set the preview.
	useEffect(() => {
		const timer = setTimeout(() => setPreview(formValue), 750);
		return () => clearTimeout(timer);
	}, [formValue]);

	return (
		<Col xs={12} sm={12} className={cx('py-1', 'col-form')}>
			<FormGroup>
				<FormLabel htmlFor={name} text={label} tooltip={tooltip} />
				<div style={{ display: 'flex', alignItems: 'center' }}>
					<FormControl
						id={name}
						aria-label={`Value field for ${label}`}
						aria-describedby={cx(
							error && `${name}-error`,
							tooltip && `${name}-tooltip`,
						)}
						type="text"
						value={formValue}
						placeholder={placeholder || defaultVal}
						style={{ marginRight: previewURL ? '1rem' : undefined }}
						autoFocus={false}
						{...register(name, {
							validate: {
								isURL: (value) => urlTest(value || defaultVal, isURL),
							},
							onBlur: () => setPreview(formValue),
						})}
						isInvalid={!!error}
					/>{' '}
					{previewURL && (
						<div
							aria-label="Preview of the image"
							style={{
								maxWidth: '3em',
								marginLeft: 'auto',
							}}
						>
							<img
								src={previewURL}
								alt="Icon preview"
								style={{ height: '2em', width: 'auto' }}
							/>
						</div>
					)}
				</div>
			</FormGroup>
			{error && (
				<small id={`${name}-error`} className="error-msg" role="alert">
					{error['message'] || 'err'}
				</small>
			)}
		</Col>
	);
};

export default FormTextWithPreview;
