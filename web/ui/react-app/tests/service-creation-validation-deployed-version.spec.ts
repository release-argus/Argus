import { expect, test } from '@playwright/test';
import { screenshotsUnder } from './fixtures/service';
import {
	clickViaKeyboard,
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
		test('type=url - URL is optional but must be a valid HTTP(S) URL if set', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/deployed-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Deployed Version');

			await section.locator('#deployed_version\\.type').click();
			await dialog.getByRole('option', { name: /^url$/i }).click();

			const urlInput = section.getByRole('textbox', {
				name: /^Value field for URL$/i,
			});

			await expectValid(urlInput, undefined, shot, '01-url-empty-ok');

			await expectError(
				urlInput,
				'example.com',
				"Invalid URL (Must start with 'http://' or 'https://').",
				shot,
				'02-url-invalid',
			);

			await expectValid(urlInput, 'https://example.com', shot, '03-url-valid');
			await expectError(
				urlInput,
				'file:///etc/passwd',
				"Invalid URL (Must start with 'http://' or 'https://').",
				shot,
				'04-url-file-invalid',
			);
		});

		test('type=file - path and capture regex are required and validated', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/deployed-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Deployed Version');

			await section.locator('#deployed_version\\.type').click();
			await dialog.getByRole('option', { name: /^file$/i }).click();

			const pathInput = section.locator(
				'input[name="deployed_version.path"]',
			);
			await expectError(
				pathInput,
				undefined,
				'Required.',
				shot,
				'01-file-path-required',
			);
			await expectError(
				pathInput,
				'relative/VERSION',
				'Must be an absolute path.',
				shot,
				'02-file-path-must-be-absolute',
			);
			await expectValid(
				pathInput,
				'/path/to/allowed/files/VERSION',
				shot,
				'03-file-path-valid',
			);

			const regexInput = section.locator(
				'input[name="deployed_version.regex"]',
			);
			await expectError(
				regexInput,
				undefined,
				'Required.',
				shot,
				'04-file-regex-required',
			);
			await expectError(
				regexInput,
				'(',
				'Invalid regular expression.',
				shot,
				'05-file-regex-invalid',
			);
			await expectError(
				regexInput,
				'productversion="[^"]+"',
				'Must include a capture group containing the version.',
				shot,
				'06-file-regex-capture-required',
			);
			await expectValid(
				regexInput,
				'productversion="([^"]+)"',
				shot,
				'07-file-regex-valid',
			);
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

			await section.locator('#deployed_version\\.type').click();
			await dialog.getByRole('option', { name: /^url$/i }).click();

			const regexInput = section.getByRole('textbox', {
				exact: true,
				name: 'Value field for RegEx',
			});

			await expectValid(
				regexInput,
				undefined,
				shot,
				'01-regex-empty-toggle-off',
			);

			const templateToggle = section.locator(
				'#deployed_version\\.template_toggle',
			);
			await templateToggle.click();

			await expectError(
				regexInput,
				undefined,
				'Required.',
				shot,
				'02-regex-required-toggle-on',
			);

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

			await section.locator('#deployed_version\\.type').click();
			await dialog.getByRole('option', { name: /^manual$/i }).click();

			const versionInput = section.locator(
				'input[name="deployed_version.version"]',
			);
			await expect(versionInput).toBeVisible();
			const versionField = fieldOf(versionInput);

			await versionInput.fill('not-a-semver-and-thats-fine');
			await versionInput.blur();
			await shot('01-manual-version-no-validation', versionField);

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

			await section.locator('#deployed_version\\.type').click();
			await dialog.getByRole('option', { name: /^url$/i }).click();
			await clickViaKeyboard(
				section.getByRole('button', { name: /add new headers/i }),
			);

			const keyInput = section.locator(
				'input[name="deployed_version.headers.0.key"]',
			);
			const valueInput = section.locator(
				'input[name="deployed_version.headers.0.value"]',
			);

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
