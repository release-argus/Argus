import { expect, test } from '@playwright/test';
import { screenshotsUnder } from './fixtures/service';
import {
	expectError,
	expectValid,
	fieldOf,
	openCreateServiceModal,
	openSection,
} from './fixtures/validation';

test.describe('Service creation modal - field validation', () => {
	// These tests never submit, so they're safe to run fully parallel.
	test.describe.configure({ mode: 'parallel' });

	test.describe('Deployed Version', () => {
		test('type=url - URL is optional but must be valid http(s):// if set', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/deployed-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Deployed Version');

			// GIVEN: `deployed_version.type` is "url".
			await section.locator('#deployed_version\\.type').click();
			await dialog.getByRole('option', { name: /^url$/i }).click();

			const urlInput = section.getByRole('textbox', {
				name: /^Value field for URL$/i,
			});

			// GIVEN: the "URL" field is focused and blurred.
			// THEN: no error is shown.
			await expectValid(urlInput, undefined, shot, '01-url-empty-ok');

			// WHEN: an invalid URL (missing http(s)://) is entered.
			// THEN: an error is shown.
			await expectError(
				urlInput,
				'example.com',
				"Invalid URL (Must start with 'http://' or 'https://').",
				shot,
				'02-url-invalid',
			);

			// WHEN: a valid HTTPS URL is entered.
			// THEN: the error clears.
			await expectValid(urlInput, 'https://example.com', shot, '03-url-valid');
		});

		test('type=url - regex becomes required when the template toggle is enabled', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/deployed-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Deployed Version');

			// GIVEN: `deployed_version.type` is "url".
			await section.locator('#deployed_version\\.type').click();
			await dialog.getByRole('option', { name: /^url$/i }).click();

			const regexInput = section.getByRole('textbox', {
				exact: true,
				name: 'Value field for RegEx',
			});

			// GIVEN: the regex field is empty and the template toggle is off.
			// THEN: no error is shown - regex is optional by default.
			await expectValid(
				regexInput,
				undefined,
				shot,
				'01-regex-empty-toggle-off',
			);

			// WHEN: the "RegEx Template" toggle is enabled.
			const templateToggle = section.locator(
				'#deployed_version\\.template_toggle',
			);
			await templateToggle.click();

			// AND: the (still-empty) regex field is blurred to trigger validation.
			// THEN: regex is now required.
			await expectError(
				regexInput,
				undefined,
				'Required.',
				shot,
				'02-regex-required-toggle-on',
			);

			// WHEN: a regex value is entered.
			// THEN: the error clears.
			await expectValid(
				regexInput,
				'v([0-9.]+)',
				shot,
				'03-regex-valid-toggle-on',
			);
		});

		test('type=manual - version field has no field-level validation', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/deployed-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Deployed Version');

			// GIVEN: `deployed_version.type` is "manual".
			await section.locator('#deployed_version\\.type').click();
			await dialog.getByRole('option', { name: /^manual$/i }).click();

			const versionInput = section.locator(
				'input[name="deployed_version.version"]',
			);
			await expect(versionInput).toBeVisible();
			const versionField = fieldOf(versionInput);

			// WHEN: an arbitrary string is entered into "Version" and blurred.
			await versionInput.fill('not-a-semver-and-thats-fine');
			await versionInput.blur();
			await shot('01-manual-version-no-validation', versionField);

			// THEN: no validation error is rendered - type=manual has no validator.
			await expect(versionInput).not.toHaveAttribute('aria-invalid', 'true');
			await expect(versionField.getByRole('alert')).not.toBeVisible();
		});

		test('type=url - header key and value are both required', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/deployed-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Deployed Version');

			// GIVEN: `deployed_version.type` is "url" with a header row added.
			await section.locator('#deployed_version\\.type').click();
			await dialog.getByRole('option', { name: /^url$/i }).click();
			await section.getByRole('button', { name: /add new headers/i }).click();

			const keyInput = section.locator(
				'input[name="deployed_version.headers.0.key"]',
			);
			const valueInput = section.locator(
				'input[name="deployed_version.headers.0.value"]',
			);

			// WHEN: the key and value are left empty and blurred.
			// THEN: a "Required." error is shown for both.
			await expectError(
				keyInput,
				undefined,
				'Required.',
				shot,
				'01-header-key-empty',
				section,
			);
			await expectError(
				valueInput,
				undefined,
				'Required.',
				shot,
				'02-header-value-empty',
				section,
			);

			// WHEN: values are entered.
			// THEN: both errors clear.
			await expectValid(
				keyInput,
				'X-Example',
				shot,
				'03-header-key-valid',
				section,
			);
			await expectValid(
				valueInput,
				'abc',
				shot,
				'04-header-value-valid',
				section,
			);
		});
	});
});
