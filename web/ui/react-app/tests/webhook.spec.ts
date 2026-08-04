import { expect, test } from '@playwright/test';
import { openDashboardInEditMode, serviceCard } from './fixtures/dashboard';
import {
	createService,
	LOOKUP_LATEST_VERSION_JSON,
	screenshot,
	trackCreatedServices,
	withProject,
} from './fixtures/service';
import { WEBHOOK_GITHUB } from './fixtures/test-endpoints';

test.describe('WebHook actions', () => {
	// Scoped to this test's ID so its afterEach can't race the other test's
	// service. IDs are project-suffixed to avoid cross-browser collisions.
	const createdIDs = trackCreatedServices();

	test('WebHook send succeeds', async ({ page }, testInfo) => {
		const baseID = 'WEBHOOK=PASS';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		// GIVEN: a service with an update available (latest_version !==
		// deployed_version) and a WebHook configured with valid credentials.
		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
			webhooks: [
				{
					name: id,
					secret: WEBHOOK_GITHUB.secretPass,
					url: WEBHOOK_GITHUB.urlValid,
				},
			],
		});
		await screenshot(
			page,
			`webhook/${baseID}/01-after-create`,
			testInfo.project.name,
		);

		const card = serviceCard(page, id);

		// WHEN: the user opens the "Resend actions" (visible text "Approve")
		// modal for the service.
		await card.getByRole('button', { name: /approve|resend/i }).click();
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await screenshot(
			page,
			`webhook/${baseID}/02-action-modal`,
			testInfo.project.name,
		);

		// AND: clicks "Send" for the WebHook.
		const sendButton = dialog.getByRole('button', {
			exact: true,
			name: 'Send',
		});
		await expect(sendButton).toBeVisible();
		await sendButton.click();

		// THEN: the WebHook send succeeds (real network call to
		// valid.release-argus.io - allow extra time).
		await expect(dialog.locator('[aria-label="Successful"]')).toBeVisible({
			timeout: 30_000,
		});
		await screenshot(
			page,
			`webhook/${baseID}/03-send-successful`,
			testInfo.project.name,
		);

		// AND: the send/retry button becomes disabled - sends are blocked until
		// the WebHook's `next_runnable` time.
		const retryButton = dialog.getByRole('button', {
			exact: true,
			name: 'Retry',
		});
		await expect(retryButton).toBeVisible();
		await expect(retryButton).toBeDisabled();

		// AND: an hourglass icon with a "Can resend ..." tooltip indicates when
		// the WebHook can next be sent.
		await dialog.locator('[aria-label="Resend timer"]').hover();
		await expect(page.getByText(/can resend/i).first()).toBeVisible();
		await screenshot(
			page,
			`webhook/${baseID}/04-send-blocked`,
			testInfo.project.name,
		);

		// WHEN: the user closes the modal.
		const doneButton = dialog.locator('#modal-action');
		await expect(doneButton).toHaveText(/^Done$/i);
		await doneButton.click();

		// THEN: the modal closes.
		await expect(dialog).not.toBeVisible();
	});

	test('WebHook send fails', async ({ page }, testInfo) => {
		const baseID = 'WEBHOOK=FAIL';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		// GIVEN: a service with an update available and a WebHook configured
		// with an invalid secret and a single try (fail fast).
		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
			webhooks: [
				{
					maxTries: '1',
					name: id,
					secret: WEBHOOK_GITHUB.secretFail,
					url: WEBHOOK_GITHUB.urlValid,
				},
			],
		});
		await screenshot(
			page,
			`webhook/${baseID}/01-after-create`,
			testInfo.project.name,
		);

		const card = serviceCard(page, id);

		// WHEN: the user opens the "Resend actions" (visible text "Approve")
		// modal for the service.
		await card.getByRole('button', { name: /approve|resend/i }).click();
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await screenshot(
			page,
			`webhook/${baseID}/02-action-modal`,
			testInfo.project.name,
		);

		// AND: clicks "Send" for the WebHook.
		const sendButton = dialog.getByRole('button', {
			exact: true,
			name: 'Send',
		});
		await expect(sendButton).toBeVisible();
		await sendButton.click();

		// THEN: the WebHook send fails (real network call - the receiver
		// responds non-2XX for an incorrect secret).
		await expect(dialog.locator('[aria-label="Failed"]')).toBeVisible({
			timeout: 30_000,
		});
		await screenshot(
			page,
			`webhook/${baseID}/03-send-failed`,
			testInfo.project.name,
		);

		// AND: the send/retry button becomes disabled - sends are blocked until
		// the WebHook's `next_runnable` time.
		const retryButton = dialog.getByRole('button', {
			exact: true,
			name: 'Retry',
		});
		await expect(retryButton).toBeVisible();
		await expect(retryButton).toBeDisabled();

		// AND: an hourglass icon with a "Can resend ..." tooltip indicates when
		// the WebHook can next be sent.
		await dialog.locator('[aria-label="Resend timer"]').hover();
		await expect(page.getByText(/can resend/i).first()).toBeVisible();
		await screenshot(
			page,
			`webhook/${baseID}/04-send-blocked`,
			testInfo.project.name,
		);

		// WHEN: the user closes the modal.
		const doneButton = dialog.locator('#modal-action');
		await expect(doneButton).toHaveText(/^Done$/i);
		await doneButton.click();

		// THEN: the modal closes.
		await expect(dialog).not.toBeVisible();
	});
});
