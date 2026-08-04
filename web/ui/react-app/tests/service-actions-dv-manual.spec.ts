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

		// GIVEN: a service with a manual deployed_version, an update available,
		// and no WebHooks/Commands.
		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});

		const card = serviceCard(page, id);

		// THEN: both the "Skip" and "Approve" actions are shown, and both the
		// deployed and latest versions are visible.
		await expect(
			card.getByRole('button', { name: 'Reject release' }),
		).toBeVisible();
		await expect(
			card.getByRole('button', { name: 'Approve release' }),
		).toBeVisible();
		await expect(card.getByText('0.0.1', { exact: true })).toBeVisible();
		await expect(card.getByText('1.2.3', { exact: true })).toBeVisible();
		await shot('01-update-available', card);

		// WHEN: the user clicks "Approve" and confirms in the resulting modal.
		const approveButton = card.getByRole('button', {
			name: 'Approve release',
		});
		await expect(approveButton).toBeEnabled();

		const dialog = page.getByRole('dialog');
		const confirmButton = dialog.locator('#modal-action');
		// Manual deployed_version updates are server-side rate-limited
		// - the modal closes immediately on confirm regardless of the
		// request's outcome, so retry the whole approve+confirm flow
		// if it lands inside that window.
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

		// THEN: the deployed version updates to the latest version, and the
		// "Skip"/"Approve" actions disappear (nothing left to action).
		await expect(card.getByText('1.2.3', { exact: true })).toBeVisible();
		await expect(
			card.getByRole('button', { name: 'Reject release' }),
		).toBeHidden();
		await expect(
			card.getByRole('button', { name: 'Approve release' }),
		).toBeHidden();
		// AND: the old deployed version is no longer shown - only the "@" line
		// for the now-matching deployed/latest version remains.
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

		// GIVEN: a service with a manual deployed_version, an update available,
		// and no WebHooks/Commands.
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
		// AND: the "Approve" action is shown in its primary (not-yet-skipped) colour.
		await expect(approveButton).toHaveClass(/bg-primary/);
		await shot('01-update-available', card);

		// WHEN: the user clicks "Skip" and confirms in the resulting modal.
		await card.getByRole('button', { name: 'Reject release' }).click();
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await shot('02-skip-modal', card);

		const confirmButton = dialog.locator('#modal-action');
		await expect(confirmButton).toHaveText(/^Skip release$/i);
		await confirmButton.click();
		await expect(dialog).not.toBeVisible();

		// THEN: the "Skip" action disappears but "Approve" remains - now in
		// its secondary colour, letting the user override the skip - and the
		// deployed version AND the (skipped) latest version both remain visible.
		await expect(
			card.getByRole('button', { name: 'Reject release' }),
		).toBeHidden();
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

		// GIVEN: a service with a manual deployed_version already matching the
		// latest version (up to date - no actions shown).
		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '1.2.3' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});

		const card = serviceCard(page, id);
		await expect(
			card.getByRole('button', { name: 'Reject release' }),
		).toBeHidden();
		await expect(
			card.getByRole('button', { name: 'Approve release' }),
		).toBeHidden();
		await shot('01-up-to-date', card);

		// WHEN: the deployed version is saved to a new version (that isn't the
		// latest version) via the edit modal's inline 'Save version' action.
		await saveDeployedVersionManual(page, id, '9.9.9');

		// THEN: the "Skip"/"Approve" actions immediately appear, reflecting the
		// no-longer-up-to-date state.
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

		// GIVEN: a service with a manual deployed_version, an update already
		// available (deployed != latest), and no WebHooks/Commands.
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
		await shot('01-update-available', card);

		// WHEN: the deployed version is saved to an older version (a downgrade,
		// still not the latest version) via the edit modal's inline
		// 'Save version' action, without reloading the page.
		await saveDeployedVersionManual(page, id, '0.0.0');

		// THEN: the "Skip"/"Approve" actions remain visible, and the displayed
		// deployed version updates to the new (older) value.
		await expect(
			card.getByRole('button', { name: 'Reject release' }),
		).toBeVisible();
		await expect(
			card.getByRole('button', { name: 'Approve release' }),
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

		// GIVEN: a service with a manual deployed_version, an update already
		// available (deployed != latest), and no WebHooks/Commands.
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
		await shot('01-update-available', card);

		// WHEN: within a single edit-modal session, the deployed version is
		// first saved to the current latest version.
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

		// THEN: the actions are hidden (up to date).
		await expect(
			card.locator('button[aria-label="Reject release"]'),
		).toBeHidden();
		await expect(
			card.locator('button[aria-label="Approve release"]'),
		).toBeHidden();

		// WHEN: without closing the modal, a different version is saved.
		await saveVersion('4.5.6');
		await page.keyboard.press('Escape');
		await expect(dialog).not.toBeVisible();

		// THEN: the actions reappear.
		await expect(
			card.getByRole('button', { name: 'Reject release' }),
		).toBeVisible();
		await expect(
			card.getByRole('button', { name: 'Approve release' }),
		).toBeVisible();
		await expect(card.getByText('4.5.6', { exact: true })).toBeVisible();
		await expect(card.getByText('1.2.3', { exact: true })).toBeVisible();
		await expect(card.getByText('0.0.1', { exact: true })).toHaveCount(0);
		await shot('02-actions-shown-again', card);
	});
});
