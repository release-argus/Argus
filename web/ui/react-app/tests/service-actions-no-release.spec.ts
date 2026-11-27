import { expect, type Locator, test } from '@playwright/test';
import { serviceCard } from './fixtures/dashboard';

/**
 * The services seeded from fixtures/noauth-config.yml whose latest_version
 * never resolves.
 */
const NO_RELEASE_SERVICES = [
	{
		/** Has a webhook, so would otherwise offer Resend. */
		id: 'playwright/no-release',
		name: 'No release yet',
	},
	{
		/** Has a manual deployed_version, so would otherwise offer Skip and Approve. */
		id: 'playwright/no-release-manual',
		name: 'No release yet (manual)',
	},
] as const;

/**
 * Asserts no release action is offered, as the API refuses them all until a
 * latest version is found.
 *
 * @param container - The service's card or table row.
 */
const expectNoReleaseActions = async (container: Locator) => {
	for (const name of ['Reject release', 'Approve release', 'Resend actions']) {
		await expect(container.getByRole('button', { name })).toHaveCount(0);
	}
};

test.describe('Service actions before a release is found', () => {
	for (const service of NO_RELEASE_SERVICES) {
		test(`${service.id}: the card offers no release actions`, async ({
			page,
		}) => {
			await page.goto('/approvals');

			const card = serviceCard(page, service.id);
			await expect(card.getByText(service.name, { exact: true })).toBeVisible();
			await expectNoReleaseActions(card);
		});

		test(`${service.id}: the table row offers no release actions`, async ({
			page,
		}) => {
			await page.goto('/approvals?view=table');

			const row = serviceCard(page, service.id);
			await expect(row.getByText(service.name, { exact: true })).toBeVisible();
			await expectNoReleaseActions(row);
		});
	}
});
