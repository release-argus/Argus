import { expect, test } from '@playwright/test';
import {
	openDashboardInEditMode,
	serviceCard,
	waitForDeployedVersionRefresh,
} from './fixtures/dashboard';
import {
	createService,
	LOOKUP_LATEST_VERSION_JSON,
	saveDeployedVersionManual,
	screenshotsUnder,
	trackCreatedServices,
	withProject,
} from './fixtures/service';
import { openSection } from './fixtures/validation';

// The manual Approve action is only offered to a service with no Webhooks or Commands.
test.describe('deployed_version=manual approve/skip actions', () => {
	const createdIDs = trackCreatedServices();

	test('Approve sets the deployed version and hides the actions', async ({
		page,
	}, testInfo) => {
		const shot = screenshotsUnder(
			page,
			testInfo.project.name,
			'service-actions-dv-manual/dashboard/approve',
		);
		const baseID = 'DV=MANUAL, APPROVE';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});

		const card = serviceCard(page, id);

		await expect(
			card.getByRole('button', { name: 'Reject release' }),
		).toBeVisible();
		await expect(
			card.getByRole('button', { name: 'Approve release' }),
		).toBeVisible();
		await expect(card.getByText('0.0.1', { exact: true })).toBeVisible();
		await expect(
			card.getByText('1.2.3', { exact: true }),
			'before approve',
		).toBeVisible();
		await shot('01-update-available', card);

		const approveButton = card.getByRole('button', {
			name: 'Approve release',
		});
		await expect(approveButton).toBeEnabled();

		const dialog = page.getByRole('dialog');
		const confirmButton = dialog.locator('#modal-action');
		// Manual deployed_version updates are rate-limited and the modal closes on confirm
		// regardless, so retry the whole approve+confirm flow.
		await expect(async () => {
			await approveButton.click();
			await expect(dialog).toBeVisible();
			await expect(confirmButton).toHaveText(/^Approve$/i);
			await shot('02-approve-modal', card);
			const [response] = await Promise.all([
				waitForDeployedVersionRefresh(page),
				confirmButton.click(),
			]);
			expect(response.ok()).toBeTruthy();
		}).toPass();
		await expect(dialog).not.toBeVisible();

		await expect(
			card.getByText('1.2.3', { exact: true }),
			'after approve',
		).toBeVisible();
		await expect(
			card.getByRole('button', { name: 'Reject release' }),
		).toBeHidden();
		await expect(
			card.getByRole('button', { name: 'Approve release' }),
		).toBeHidden();
		await expect(card.getByText('0.0.1', { exact: true })).toHaveCount(0);
		await shot('03-approved', card);
	});

	test('Skip hides Skip but keeps Approve (as secondary) and both versions displayed', async ({
		page,
	}, testInfo) => {
		const shot = screenshotsUnder(
			page,
			testInfo.project.name,
			'service-actions-dv-manual/dashboard/skip',
		);
		const baseID = 'dv=MANUAL, SKIP';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});

		const card = serviceCard(page, id);
		const approveButton = card.getByRole('button', {
			name: 'Approve release',
		});
		await expect(
			card.getByRole('button', { name: 'Reject release' }),
		).toBeVisible();
		// The primary colour marks a release not yet skipped.
		await expect(approveButton).toHaveClass(/bg-primary/);
		await shot('01-update-available', card);

		await card.getByRole('button', { name: 'Reject release' }).click();
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await shot('02-skip-modal', card);

		const confirmButton = dialog.locator('#modal-action');
		await expect(confirmButton).toHaveText(/^Skip release$/i);
		await confirmButton.click();
		await expect(dialog).not.toBeVisible();

		await expect(
			card.getByRole('button', { name: 'Reject release' }),
		).toBeHidden();
		// Approve stays, in secondary, so the skip can be overridden.
		await expect(approveButton).toBeVisible();
		await expect(approveButton).toHaveClass(/bg-secondary/);
		await expect(card.getByText('0.0.1', { exact: true })).toBeVisible();
		await expect(card.getByText('1.2.3', { exact: true })).toBeVisible();
		await shot('03-skipped', card);
	});

	test('Saving a new deployed version from the edit modal shows the actions', async ({
		page,
	}, testInfo) => {
		const shot = screenshotsUnder(
			page,
			testInfo.project.name,
			'service-actions-dv-manual/edit/save-new-version',
		);
		const baseID = 'DV=MANUAL, EDIT-SAVE NEW';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '1.2.3' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});

		const card = serviceCard(page, id);
		// The manual deployed_version already matches the latest version, so no
		// actions are shown.
		await expect(
			card.getByRole('button', { name: 'Reject release' }),
		).toBeHidden();
		await expect(
			card.getByRole('button', { name: 'Approve release' }),
		).toBeHidden();
		await shot('01-up-to-date', card);

		await saveDeployedVersionManual(page, id, '9.9.9');

		// No reload, so the card must reflect the save on its own.
		await expect(
			card.getByRole('button', { name: 'Reject release' }),
		).toBeVisible();
		await expect(
			card.getByRole('button', { name: 'Approve release' }),
		).toBeVisible();
		await expect(card.getByText('9.9.9', { exact: true })).toBeVisible();
		await expect(card.getByText('1.2.3', { exact: true })).toBeVisible();
		await shot('02-actions-shown', card);
	});

	test('Saving another deployed version that is not the latest version from the edit modal keeps the actions shown', async ({
		page,
	}, testInfo) => {
		const shot = screenshotsUnder(
			page,
			testInfo.project.name,
			'service-actions-dv-manual/edit/save-other-version',
		);
		const baseID = 'DV=MANUAL, EDIT-SAVE OTHER';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});

		const card = serviceCard(page, id);
		await expect(
			card.getByRole('button', { name: 'Reject release' }),
			'before save',
		).toBeVisible();
		await expect(
			card.getByRole('button', { name: 'Approve release' }),
			'before save',
		).toBeVisible();
		await shot('01-update-available', card);

		// 0.0.0 is a downgrade, so still not the latest version.
		await saveDeployedVersionManual(page, id, '0.0.0');

		// No reload, so the card must reflect the save on its own.
		await expect(
			card.getByRole('button', { name: 'Reject release' }),
			'after save',
		).toBeVisible();
		await expect(
			card.getByRole('button', { name: 'Approve release' }),
			'after save',
		).toBeVisible();
		await expect(card.getByText('0.0.0', { exact: true })).toBeVisible();
		await expect(card.getByText('1.2.3', { exact: true })).toBeVisible();
		await expect(card.getByText('0.0.1', { exact: true })).toHaveCount(0);
		await shot('02-actions-still-shown', card);
	});

	test('Saving to the latest version then to another version in the same edit session updates the actions each time', async ({
		page,
	}, testInfo) => {
		const shot = screenshotsUnder(
			page,
			testInfo.project.name,
			'service-actions-dv-manual/edit/save-latest-then-other-version',
		);
		const baseID = 'DV=MANUAL, EDIT-SAVE LATEST-THEN-OTHER';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});

		const card = serviceCard(page, id);
		await expect(
			card.getByRole('button', { name: 'Reject release' }),
			'before save',
		).toBeVisible();
		await expect(
			card.getByRole('button', { name: 'Approve release' }),
			'before save',
		).toBeVisible();
		await shot('01-update-available', card);

		await card.getByRole('button', { name: /edit/i }).click();
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		const section = await openSection(dialog, 'Deployed Version');
		const versionInput = section.locator(
			'input[name="deployed_version.version"]',
		);
		const saveButton = section.getByRole('button', { name: 'Save version' });
		const saveVersion = (version: string) =>
			expect(async () => {
				await versionInput.fill(version);
				const [response] = await Promise.all([
					waitForDeployedVersionRefresh(page),
					saveButton.click(),
				]);
				expect(response.ok()).toBeTruthy();
			}).toPass();
		await saveVersion('1.2.3');

		// The open modal aria-hides the card, which a plain role query would skip and pass vacuously.
		await expect(
			card.getByRole('button', { includeHidden: true, name: 'Reject release' }),
		).toBeHidden();
		await expect(
			card.getByRole('button', {
				includeHidden: true,
				name: 'Approve release',
			}),
		).toBeHidden();

		await saveVersion('4.5.6');
		await page.keyboard.press('Escape');
		await expect(dialog).not.toBeVisible();

		await expect(
			card.getByRole('button', { name: 'Reject release' }),
			'after second save',
		).toBeVisible();
		await expect(
			card.getByRole('button', { name: 'Approve release' }),
			'after second save',
		).toBeVisible();
		await expect(card.getByText('4.5.6', { exact: true })).toBeVisible();
		await expect(card.getByText('1.2.3', { exact: true })).toBeVisible();
		await expect(card.getByText('0.0.1', { exact: true })).toHaveCount(0);
		await shot('02-actions-shown-again', card);
	});
});
