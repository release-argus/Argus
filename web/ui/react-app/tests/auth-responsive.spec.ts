import { type BrowserContext, expect, type Page, test } from '@playwright/test';
import { ADMIN_PASSWORD, ADMIN_USER, signedInContext } from './fixtures/auth';

/**
 * Responsive-layout e2e specs for the admin UI - run only in the
 * `chromium-auth` project, against the auth-enabled Argus instance
 * (AUTH_BASE_URL, bootstrap admin created by `make playwright-tests-setup`).
 *
 * Scope: the grant editor's `sm` breakpoint. Below it each row stacks over two
 * lines and every control carries its own label; at `sm` and up the labels are
 * dropped for one shared column header. Every other auth spec runs at a desktop
 * viewport, so nothing else exercises the stacked half.
 *
 * Read-only: the group dialog is opened and dismissed without saving, so these
 * share the auth instance without mutating it.
 */

const MOBILE = { height: 844, width: 390 };
const DESKTOP = { height: 900, width: 1280 };

test.describe.configure({ mode: 'serial' });

test.describe('Admin UI responsive layout', () => {
	let context: BrowserContext;
	let page: Page;

	test.beforeAll(async ({ browser }) => {
		context = await signedInContext(browser, ADMIN_USER, ADMIN_PASSWORD);
		page = await context.newPage();
	});

	test.afterAll(async () => {
		await context?.close();
	});

	const dialog = () => page.getByRole('dialog');

	/** A control's own label, rendered only while the row is stacked. */
	const cellLabel = (text: string) =>
		dialog().locator('[data-slot="field"] > span').filter({ hasText: text });

	/**
	 * The shared column header: the one direct child of the editor that is not
	 * a grant row.
	 */
	const columnHeader = () =>
		dialog().locator('fieldset > div:not([data-slot="field-set"])');

	/** Opens the create-group dialog carrying a single grant row. */
	const openGrantEditor = async () => {
		await page.goto('/admin/groups');
		await page.getByRole('button', { name: 'Add group' }).click();
		await page.getByRole('button', { name: 'Add permission' }).click();
		await expect(page.getByLabel('Grant 0 resource')).toBeVisible();
	};

	test('stacked/labels replace the column header below sm', async () => {
		// GIVEN: a narrow viewport.
		await page.setViewportSize(MOBILE);

		// WHEN: the grant editor is opened.
		await openGrantEditor();

		// THEN: every control carries its own label, and the header is dropped.
		for (const label of ['Resource', 'Action', 'Scope']) {
			await expect(cellLabel(label)).toBeVisible();
		}
		await expect(columnHeader()).toBeHidden();
	});

	test('stacked/a scoped grant reveals an editable target', async () => {
		// GIVEN: the grant editor at a narrow viewport.
		await page.setViewportSize(MOBILE);
		await openGrantEditor();

		// AND: a global grant has no target field.
		await expect(page.getByLabel('Grant 0 scope ref')).toHaveCount(0);

		// WHEN: the grant is scoped to a service.
		await page.getByLabel('Grant 0 scope').click();
		await page.getByRole('option', { exact: true, name: 'service' }).click();

		// THEN: the target appears on the stacked row and accepts input.
		const target = page.getByLabel('Grant 0 scope ref');
		await expect(target).toBeVisible();
		await expect(cellLabel('Target')).toBeVisible();
		await target.fill('release-argus/Argus');
		await expect(target).toHaveValue('release-argus/Argus');
	});

	test('wide/the column header replaces the per-control labels', async () => {
		// GIVEN: a desktop viewport.
		await page.setViewportSize(DESKTOP);

		// WHEN: the grant editor is opened.
		await openGrantEditor();

		// THEN: the shared header is used and the per-control labels are dropped.
		await expect(columnHeader()).toBeVisible();
		for (const label of ['Resource', 'Action', 'Scope']) {
			await expect(cellLabel(label)).toBeHidden();
		}
	});
});
