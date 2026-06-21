import { expect, type Locator, type Page, test } from '@playwright/test';
import {
	type CreateServiceOptions,
	cleanupServices,
	createService,
	screenshot,
} from './fixtures/service';
import { bareEndpoint } from './fixtures/test-endpoints';

// Creating a service makes a real server-side network call that can exceed the
// default timeout under load - so triple it.
test.beforeEach(() => {
	test.slow();
});

/**
 * Reads the current service order from the rendered cards/rows.
 *
 * @param page - The dashboard page.
 * @returns The service IDs in their rendered order.
 */
const getServiceOrder = async (page: Page) => {
	await expect(page.locator('[data-service-id]').first()).toBeVisible();
	return page
		.locator('[data-service-id]')
		.evaluateAll((nodes) =>
			nodes.map((n) => n.getAttribute('data-service-id')),
		);
};

/**
 * Drags `source` onto `target`. dnd-kit needs the drag to begin with a small
 * movement (a single jump can be missed) and the pointer to cross the target's
 * centre to swap - so: nudge to activate, step across, overshoot the centre,
 * then dwell before releasing. Works for table (vertical) and grid (horizontal).
 *
 * @param page - The dashboard page.
 * @param source - The drag handle to pick up.
 * @param target - The drag handle to drop onto.
 */
const dragAndDrop = async (page: Page, source: Locator, target: Locator) => {
	const sourceBox = await source.boundingBox();
	const targetBox = await target.boundingBox();
	if (!sourceBox || !targetBox)
		throw new Error('Missing bounding box for drag');

	const fromX = sourceBox.x + sourceBox.width / 2;
	const fromY = sourceBox.y + sourceBox.height / 2;
	const toX = targetBox.x + targetBox.width / 2;
	const toY = targetBox.y + targetBox.height / 2;

	// A point one target-half past its centre, along the source->target vector.
	const dx = toX - fromX;
	const dy = toY - fromY;
	const len = Math.hypot(dx, dy) || 1;
	const pastX = toX + (dx / len) * (targetBox.width / 2 + 4);
	const pastY = toY + (dy / len) * (targetBox.height / 2 + 4);

	await page.mouse.move(fromX, fromY);
	await page.mouse.down();
	await page.mouse.move(fromX, fromY + 10, { steps: 5 }); // activate the sensor
	await page.mouse.move(toX, toY, { steps: 15 });
	await page.mouse.move(pastX, pastY, { steps: 5 }); // overshoot the centre
	await page.mouse.move(pastX, pastY); // dwell on the drop point
	await page.mouse.up();
};

/**
 * Clicks "Save order" and waits for it to persist. The short pause lets dnd-kit
 * settle - it swallows a click landing immediately after a drop.
 *
 * @param page - The dashboard page.
 */
const saveOrder = async (page: Page) => {
	await page.waitForTimeout(100);
	await page.getByRole('button', { name: /save order/i }).click();
	await expect(
		page.getByRole('button', { name: /save order/i }),
	).not.toBeVisible();
};

// `latest_version` options shared by the ordering test services.
const ORDERING_SERVICE_OPTIONS: CreateServiceOptions = {
	latestVersion: {
		allowInvalidCerts: false,
		type: 'url',
		url: bareEndpoint('1.2.3'),
	},
};

test.describe('Service Ordering', () => {
	// Both tests reuse the same two IDs, so they must run serially within a
	// project.
	test.describe.configure({ mode: 'serial' });

	// Reordering writes the *global* dashboard order shared by every browser
	// project, so running on a single browser avoids the cross-project race
	// (the drag is exercised here; persistence is browser-agnostic).
	test.skip(
		({ browserName }) => browserName !== 'chromium',
		'reorders shared global state - runs on one project only',
	);

	const ids = (projectName: string) => ({
		one: `service-1-${projectName}`,
		two: `service-2-${projectName}`,
	});

	test.afterEach(async ({ page }, testInfo) => {
		const { one, two } = ids(testInfo.project.name);
		await cleanupServices(page, [one, two]);
	});

	test('re-order in grid view', async ({ page }, testInfo) => {
		const { one, two } = ids(testInfo.project.name);

		await page.goto('/');
		await page.getByRole('button', { name: /toggle edit mode/i }).click();
		await createService(page, one, ORDERING_SERVICE_OPTIONS);
		await createService(page, two, ORDERING_SERVICE_OPTIONS);
		await screenshot(
			page,
			'service-ordering/grid/01-after-creates',
			testInfo.project.name,
		);

		await page.getByRole('radio', { name: 'Grid layout' }).click();

		const dragHandle1Grid = page
			.locator(`[data-service-id="${one}"]`)
			.getByRole('button', { name: /drag handle/i });
		const targetCard2Grid = page
			.locator(`[data-service-id="${two}"]`)
			.getByRole('button', { name: /drag handle/i });
		await expect(dragHandle1Grid).toBeVisible();
		await expect(targetCard2Grid).toBeVisible();
		await dragAndDrop(page, dragHandle1Grid, targetCard2Grid);

		// dnd-kit applies the reorder on a later frame, so the rendered order can
		// briefly lag the drop - poll until it reflects the swap.
		await expect(async () => {
			const order = await getServiceOrder(page);
			expect(order.indexOf(two)).toBeLessThan(order.indexOf(one));
		}).toPass({ timeout: 10_000 });

		await saveOrder(page);
		await screenshot(
			page,
			'service-ordering/grid/02-after-reorder',
			testInfo.project.name,
		);

		await page.reload();
		const orderAfterReload = await getServiceOrder(page);
		expect(orderAfterReload.indexOf(two)).toBeLessThan(
			orderAfterReload.indexOf(one),
		);
		await screenshot(
			page,
			'service-ordering/grid/03-after-refresh',
			testInfo.project.name,
		);
	});

	test('re-order in table view', async ({ page }, testInfo) => {
		const { one, two } = ids(testInfo.project.name);

		await page.goto('/');
		await page.getByRole('button', { name: /toggle edit mode/i }).click();
		await createService(page, one, ORDERING_SERVICE_OPTIONS);
		await createService(page, two, ORDERING_SERVICE_OPTIONS);
		await screenshot(
			page,
			'service-ordering/table/01-after-creates',
			testInfo.project.name,
		);

		await page.getByRole('radio', { name: 'Table layout' }).click();
		await expect(
			page.getByRole('radio', { name: 'Table layout' }),
		).toBeChecked();

		const dragHandle1Table = page
			.locator(`[data-service-id="${one}"]`)
			.getByRole('button', { name: /drag handle/i });
		const targetCard2Table = page
			.locator(`[data-service-id="${two}"]`)
			.getByRole('button', { name: /drag handle/i });
		await expect(dragHandle1Table).toBeVisible();
		await expect(targetCard2Table).toBeVisible();
		await dragAndDrop(page, dragHandle1Table, targetCard2Table);

		// dnd-kit applies the reorder on a later frame, so the rendered order can
		// briefly lag the drop - poll until it reflects the swap.
		await expect(async () => {
			const order = await getServiceOrder(page);
			expect(order.indexOf(two)).toBeLessThan(order.indexOf(one));
		}).toPass({ timeout: 10_000 });

		await saveOrder(page);
		await screenshot(
			page,
			'service-ordering/table/02-after-reorder',
			testInfo.project.name,
		);

		await page.reload();
		const orderAfterReload = await getServiceOrder(page);
		expect(orderAfterReload.indexOf(two)).toBeLessThan(
			orderAfterReload.indexOf(one),
		);
		await screenshot(
			page,
			'service-ordering/table/03-after-refresh',
			testInfo.project.name,
		);
	});
});
