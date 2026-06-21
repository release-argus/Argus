import { expect, test } from '@playwright/test';
import { screenshotsUnder } from './fixtures/service';
import {
	expectError,
	expectValid,
	openCreateServiceModal,
	openSection,
	valueInputFor,
} from './fixtures/validation';

test.describe('Service creation modal - field validation', () => {
	// These tests never submit, so they're safe to run fully parallel.
	test.describe.configure({ mode: 'parallel' });

	test.describe('Options', () => {
		test('interval must be a valid AhBmCs duration', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/options',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Options');

			const intervalInput = valueInputFor(section, 'Interval');

			// WHEN: an invalid duration is entered and blurred.
			// THEN: an error is shown.
			await expectError(
				intervalInput,
				'abc',
				"Invalid duration. Use 'AhBmCs' duration format.",
				shot,
				'01-invalid-interval',
				section,
			);

			// WHEN: the value is corrected to a valid duration.
			// THEN: the error clears.
			await expectValid(intervalInput, '10m', shot, '02-valid-interval');
		});
	});

	test.describe('Command', () => {
		test('argument is required', async ({ page }, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/command',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Command:');

			// GIVEN: a new command is added (with one empty argument).
			await section.getByRole('button', { name: /add command/i }).click();

			const argInput = section.locator('input[name="command.0.0.arg"]');
			await expect(argInput).toBeVisible();

			// WHEN: the argument is left empty and blurred.
			// THEN: an error is shown.
			await expectError(
				argInput,
				'',
				'Required.',
				shot,
				'01-arg-required',
				section,
			);

			// WHEN: a value is entered.
			// THEN: the error clears.
			await expectValid(argInput, '/bin/bash', shot, '02-arg-valid');
		});
	});

	test.describe('Service ID / Name', () => {
		// The seeded service in `config.yml.example`, reused as the existing
		// ID/Name to collide against. Matched case-sensitively, so this must be
		// the exact seeded id.
		const EXISTING_ID = 'release-argus/Argus';

		test('ID and Name must be unique against existing services', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/service',
			);
			const dialog = await openCreateServiceModal(page);

			const idInput = dialog.locator('input[name="id"]');

			// WHEN: an existing service's ID is entered and blurred.
			// THEN: an error is shown.
			await expectError(
				idInput,
				EXISTING_ID,
				'Must be unique.',
				shot,
				'01-id-duplicate',
			);

			// WHEN: a unique ID is entered.
			// THEN: the error clears.
			await expectValid(
				idInput,
				'VALIDATION_TEST_unique',
				shot,
				'02-id-unique',
			);

			// WHEN: ID and Name are separated, revealing a distinct "Name" field.
			await dialog
				.getByRole('button', { name: /toggle to separate id/i })
				.click();
			const nameInput = dialog.locator('input[name="name"]');

			// AND: an existing service's ID is entered as the Name and blurred.
			// THEN: an error is shown.
			await expectError(
				nameInput,
				EXISTING_ID,
				'Must be unique.',
				shot,
				'03-name-duplicate',
			);

			// WHEN: a unique Name is entered.
			// THEN: the error clears.
			await expectValid(
				nameInput,
				'VALIDATION_TEST_unique_name',
				shot,
				'04-name-unique',
			);
		});
	});

	test.describe('Dashboard', () => {
		test('icon, web URL, and icon link fields have no field-level validation', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/dashboard',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Dashboard:');

			const iconInput = section.locator('input[name="dashboard.icon"]');
			const webUrlInput = section.locator('input[name="dashboard.web_url"]');
			const iconLinkInput = section.locator(
				'input[name="dashboard.icon_link_to"]',
			);

			// WHEN: arbitrary, non-URL strings are entered into "Icon", "Web URL",
			// and "Icon link to", and each is blurred.
			await iconInput.fill('not-a-url');
			await iconInput.blur();
			await webUrlInput.fill('not-a-url-either');
			await webUrlInput.blur();
			await iconLinkInput.fill('also-not-a-url');
			await iconLinkInput.blur();
			await shot('01-arbitrary-values');

			// THEN: no validation errors are rendered - these fields have no validator.
			await expect(iconInput).not.toHaveAttribute('aria-invalid', 'true');
			await expect(webUrlInput).not.toHaveAttribute('aria-invalid', 'true');
			await expect(iconLinkInput).not.toHaveAttribute('aria-invalid', 'true');
			await expect(section.getByRole('alert')).toHaveCount(0);
		});
	});
});
