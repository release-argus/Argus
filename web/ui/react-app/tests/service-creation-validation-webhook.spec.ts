import { expect, test } from '@playwright/test';
import { screenshotsUnder } from './fixtures/service';
import {
	expectError,
	fieldOf,
	MUST_BE_UNIQUE,
	NUMBER_REQUIRED,
	openCreateServiceModal,
	openSection,
	REQUIRED,
	required,
	runValidations,
} from './fixtures/validation';

test.describe('Service creation modal - field validation', () => {
	// These tests never submit, so they're safe to run fully parallel.
	test.describe.configure({ mode: 'parallel' });

	test.describe('WebHook', () => {
		test('name, target URL, and secret are required strings; max tries must be a valid in-range number', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/webhook',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'WebHook:');

			// GIVEN: a new webhook is added.
			await section.getByRole('button', { name: /add webhook/i }).click();
			const header = section.locator('[data-slot="accordion-trigger"]', {
				hasText: /^0:/,
			});
			await expect(header).toBeVisible();
			await header.click();

			// WHEN/THEN: name/URL/secret are blurred empty then valid.
			// AND: max tries rejects a non-number, then an out-of-range value,
			// then accepts a valid one.
			await runValidations(section, shot, [
				required({ good: 'my-webhook', input: 'Name', slug: 'name' }),
				required({
					good: 'https://example.com/hook',
					input: 'Target URL',
					slug: 'url',
				}),
				required({ good: 's3cret', input: 'Secret', slug: 'secret' }),
				{
					bad: 'abc',
					badSlug: 'max-tries-not-a-number',
					error: NUMBER_REQUIRED,
					input: 'Max tries',
				},
				{
					bad: '999',
					badSlug: 'max-tries-out-of-range',
					error: 'Must be between 0 and 255.',
					good: '3',
					goodSlug: 'max-tries-valid',
					input: 'Max tries',
				},
			]);
		});

		test('type=gitlab - same Name/URL/Secret validation applies after switching type', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/webhook',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'WebHook:');

			// GIVEN: a new webhook is added (defaults to type "github").
			await section.getByRole('button', { name: /add webhook/i }).click();
			const header = section.locator('[data-slot="accordion-trigger"]', {
				hasText: /^0:/,
			});
			await expect(header).toBeVisible();
			await header.click();

			// WHEN: its type is switched to "GitLab".
			await section.locator('#webhook\\.0\\.type').click();
			await dialog.getByRole('option', { name: /^gitlab$/i }).click();
			await expect(section.locator('#webhook\\.0\\.type')).toHaveText(
				/gitlab/i,
			);
			await shot('01-gitlab-type');

			// AND: "Name" is cleared and blurred.
			const nameInput = section.getByRole('textbox', {
				name: 'Value field for Name',
			});
			const nameField = fieldOf(nameInput);
			await nameInput.fill('');
			await nameInput.blur();

			// THEN: the same error still applies - type (github/gitlab) doesn't
			// affect validation.
			await expect(nameInput).toHaveAttribute('aria-invalid', 'true');
			await expect(nameField.getByRole('alert')).toHaveText(REQUIRED);

			// WHEN: a valid value is entered.
			await nameInput.fill('my-gitlab-webhook');
			await nameInput.blur();

			// THEN: the error clears.
			await expect(nameInput).toHaveAttribute('aria-invalid', 'false');
			await expect(nameField.getByRole('alert')).not.toBeVisible();
		});

		test('duplicate webhook names are flagged as not unique', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/webhook',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'WebHook:');
			const addWebHook = section.getByRole('button', { name: /add webhook/i });

			// GIVEN: a first webhook is added, expanded, and named "dup".
			await addWebHook.click();
			const header0 = section.locator('[data-slot="accordion-trigger"]', {
				hasText: /^0:/,
			});
			await expect(header0).toBeVisible();
			await header0.click();
			// Both webhooks share the "Value field for Name" textbox name, so
			// index them by position rather than an attribute locator.
			const nameInputs = section.getByRole('textbox', {
				name: 'Value field for Name',
			});
			const name0 = nameInputs.nth(0);
			const name0Field = fieldOf(name0);
			await name0.fill('dup');
			await name0.blur();

			// AND: a second webhook is added and expanded.
			await addWebHook.click();
			const header1 = section.locator('[data-slot="accordion-trigger"]', {
				hasText: /^1:/,
			});
			await expect(header1).toBeVisible();
			await header1.click();
			const name1 = nameInputs.nth(1);
			const name1Field = fieldOf(name1);

			// WHEN: the second webhook is given the first's name and blurred.
			// THEN: it is flagged as not unique.
			await expectError(
				name1,
				'dup',
				MUST_BE_UNIQUE,
				shot,
				'01-webhook-name-duplicate',
			);

			// AND: the first webhook is flagged too - the uniqueness check flags
			// both names; re-blur it to surface the now-applicable error.
			await name0.focus();
			await name0.blur();
			await expect(name0).toHaveAttribute('aria-invalid', 'true');
			await expect(name0Field.getByRole('alert')).toHaveText(MUST_BE_UNIQUE);

			// WHEN: the second name is made unique. THEN: both errors clear.
			await name1.fill('dup-2');
			await name1.blur();
			await expect(name1).toHaveAttribute('aria-invalid', 'false');
			await expect(name1Field.getByRole('alert')).not.toBeVisible();
			await name0.focus();
			await name0.blur();
			await expect(name0).toHaveAttribute('aria-invalid', 'false');
			await expect(name0Field.getByRole('alert')).not.toBeVisible();
		});
	});
});
