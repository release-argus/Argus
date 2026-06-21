import { expect, test } from '@playwright/test';
import { screenshotsUnder } from './fixtures/service';
import {
	expectError,
	expectValid,
	openCreateServiceModal,
	openSection,
} from './fixtures/validation';

test.describe('Service creation modal - field validation', () => {
	// These tests never submit, so they're safe to run fully parallel.
	test.describe.configure({ mode: 'parallel' });

	test.describe('Latest Version', () => {
		test('type=github - repository is required and must match owner/repo', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			// GIVEN: `latest_version.type` is "github" (default).
			const repoInput = section.getByRole('textbox', { name: /repository/i });

			// GIVEN: the "Repository" field is focused and blurred.
			// THEN: a "Required." error is shown.
			await expectError(
				repoInput,
				undefined,
				'Required.',
				shot,
				'01-github-repo-empty',
			);

			// WHEN: an invalid "owner/repo" value is entered.
			// THEN: an error is shown.
			await expectError(
				repoInput,
				'not-a-repo',
				'Invalid GitHub repository.',
				shot,
				'02-github-repo-invalid',
			);

			// WHEN: a valid "owner/repo" value is entered.
			// THEN: the error clears.
			await expectValid(
				repoInput,
				'release-argus/Argus',
				shot,
				'03-github-repo-valid',
			);
		});

		test('type=url - URL is required and must start with http(s)://', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			// GIVEN: `latest_version.type` is "url".
			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^url$/i }).click();

			const urlInput = section.getByRole('textbox', {
				name: /^Value field for URL$/i,
			});

			// GIVEN: the "URL" field is focused and blurred.
			// THEN: an error is shown.
			await expectError(urlInput, undefined, 'Required.', shot, '01-url-empty');

			// WHEN: a valid URL is entered.
			// THEN: no error is present.
			await expectValid(urlInput, 'https://example.com', shot, '02-url-valid');

			// WHEN: an invalid URL (missing http(s)://) is entered.
			// THEN: an error is shown.
			await expectError(
				urlInput,
				'example.com',
				"Invalid URL (Must start with 'http://' or 'https://').",
				shot,
				'03-url-invalid',
			);
		});

		test('url_commands - regex command must be a valid regular expression', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			// GIVEN: `latest_version.type` is "url" and we've added a url command.
			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^url$/i }).click();
			await section
				.getByRole('button', { name: /add new url command/i })
				.click();

			const regexInput = section.locator(
				'input[name="latest_version.url_commands.0.regex"]',
			);
			await expect(regexInput).toBeVisible();

			// WHEN: an unbalanced (invalid) regular expression is entered.
			// THEN: an error is shown.
			await expectError(
				regexInput,
				'(',
				'Invalid regular expression.',
				shot,
				'01-url-command-regex-invalid',
			);

			// WHEN: the regex is corrected to a valid expression.
			// THEN: the error clears.
			await expectValid(
				regexInput,
				'v([0-9.]+)',
				shot,
				'02-url-command-regex-valid',
			);

			// WHEN: the regex field is emptied.
			// THEN: an error is shown.
			await expectError(
				regexInput,
				'',
				'Required.',
				shot,
				'03-url-command-regex-required',
			);
		});

		test('url_commands - replace and split commands require their fields', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			// GIVEN: `latest_version.type` is "url" with a url command added.
			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^url$/i }).click();
			await section
				.getByRole('button', { name: /add new url command/i })
				.click();

			// WHEN: the command type is switched to "Replace".
			// THEN: its "Replace" (`old`) field is required.
			await section
				.locator('#latest_version\\.url_commands\\.0\\.type')
				.click();
			await dialog.getByRole('option', { name: /^replace$/i }).click();
			const oldInput = section.locator(
				'input[name="latest_version.url_commands.0.old"]',
			);
			await expectError(
				oldInput,
				undefined,
				'Required.',
				shot,
				'01-url-command-replace-old-empty',
				section,
			);
			await expectValid(
				oldInput,
				'beta',
				shot,
				'02-url-command-replace-old-valid',
				section,
			);

			// WHEN: the command type is switched to "Split".
			// THEN: both its "Text" and "Index" fields are required.
			await section
				.locator('#latest_version\\.url_commands\\.0\\.type')
				.click();
			await dialog.getByRole('option', { name: /^split$/i }).click();
			const textInput = section.locator(
				'input[name="latest_version.url_commands.0.text"]',
			);
			await expectError(
				textInput,
				undefined,
				'Required.',
				shot,
				'03-url-command-split-text-empty',
				section,
			);
			await expectValid(
				textInput,
				'-',
				shot,
				'04-url-command-split-text-valid',
				section,
			);
			const indexInput = section.locator(
				'input[name="latest_version.url_commands.0.index"]',
			);
			await expectError(
				indexInput,
				undefined,
				'Required.',
				shot,
				'05-url-command-split-index-empty',
				section,
			);
			await expectValid(
				indexInput,
				'0',
				shot,
				'06-url-command-split-index-valid',
				section,
			);
		});

		test('require - regex_content and regex_version must be valid regular expressions', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			// GIVEN: the "Require:" sub-accordion is expanded.
			await section.getByRole('button', { name: /^require:?$/i }).click();

			// Both regex fields are optional (empty is fine) but must compile to a
			// valid RegExp when set.
			const contentInput = section.locator(
				'input[name="latest_version.require.regex_content"]',
			);
			await expectValid(
				contentInput,
				undefined,
				shot,
				'01-require-regex-content-empty',
				section,
			);
			await expectError(
				contentInput,
				'(',
				'Invalid regular expression.',
				shot,
				'02-require-regex-content-invalid',
				section,
			);
			await expectValid(
				contentInput,
				'v([0-9.]+)',
				shot,
				'03-require-regex-content-valid',
				section,
			);

			const versionInput = section.locator(
				'input[name="latest_version.require.regex_version"]',
			);
			await expectValid(
				versionInput,
				undefined,
				shot,
				'04-require-regex-version-empty',
				section,
			);
			await expectError(
				versionInput,
				'[',
				'Invalid regular expression.',
				shot,
				'05-require-regex-version-invalid',
				section,
			);
			await expectValid(
				versionInput,
				'^[0-9.]+$',
				shot,
				'06-require-regex-version-valid',
				section,
			);
		});

		test('require.docker - image and tag must be set together', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			// GIVEN: the "Require:" sub-accordion is expanded.
			await section.getByRole('button', { name: /^require:?$/i }).click();

			const imageInput = section.locator(
				'input[name="latest_version.require.docker.image"]',
			);
			const tagInput = section.locator(
				'input[name="latest_version.require.docker.tag"]',
			);

			// GIVEN: both image and tag are set (a satisfied pairing - no error).
			await imageInput.fill('release-argus/argus');
			await imageInput.blur();
			await tagInput.fill('latest');
			await tagInput.blur();

			// WHEN: the image is cleared while the tag remains.
			// THEN: the image becomes required - docker needs image and tag together
			// (or neither).
			await expectError(
				imageInput,
				'',
				'Required.',
				shot,
				'01-require-docker-image-required',
				section,
			);

			// WHEN: the image is restored.
			// THEN: the pairing is satisfied, the error clears.
			await expectValid(
				imageInput,
				'release-argus/argus',
				shot,
				'02-require-docker-paired-valid',
				section,
			);
		});
	});
});
