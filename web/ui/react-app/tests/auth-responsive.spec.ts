import { type BrowserContext, expect, type Page, test } from '@playwright/test';
import { ADMIN_PASSWORD, ADMIN_USER, signedInContext } from './fixtures/auth';

/**
 * Responsive-layout e2e specs for the admin UI - run only in the
 * `chromium-auth` project, against the auth-enabled Argus instance
 * (AUTH_BASE_URL, bootstrap admin created by `make playwright-tests-setup`).
 *
 * Scope: the two things that shift with the viewport. The grant editor's `sm`
 * breakpoint, and the on-screen keyboard, which takes its height off the
 * viewport.
 *
 * Read-only: the group dialog is opened and dismissed without saving, so these
 * share the auth instance without mutating it.
 */

const MOBILE = { height: 844, width: 390 };
const DESKTOP = { height: 900, width: 1280 };

/** A phone with the keyboard up; `resizes-content` takes it off the viewport. */
const MOBILE_KEYBOARD = { height: 367, width: 390 };

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
		await expect(page.getByLabel('Grant 0 target')).toHaveCount(0);

		// WHEN: the grant is scoped to a service.
		await page.getByLabel('Grant 0 scope').click();
		await page.getByRole('option', { exact: true, name: 'service' }).click();

		// THEN: the target appears on the stacked row and accepts input.
		const target = page.getByLabel('Grant 0 target');
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

	test('the viewport opts into keyboard resizing', async () => {
		// GIVEN: any page of the app.
		await page.goto('/admin/users');

		// WHEN/THEN: the viewport meta opts in.
		await expect(page.locator('meta[name="viewport"]')).toHaveAttribute(
			'content',
			/interactive-widget=resizes-content/,
		);
	});

	test('a dialog stays reachable under the keyboard', async () => {
		// GIVEN: a phone viewport with the keyboard up.
		await page.setViewportSize(MOBILE_KEYBOARD);

		// WHEN: the create-user dialog is opened.
		await page.goto('/admin/users');
		await page.getByRole('button', { name: 'Add user' }).click();
		await expect(dialog()).toBeVisible();

		// THEN: it fits within the viewport.
		const bounds = await page.evaluate(() => {
			const box = document
				.querySelector('[role=dialog]')
				?.getBoundingClientRect();
			if (!box) throw new Error('no dialog');
			return {
				bottom: box.bottom <= globalThis.innerHeight,
				top: box.top >= 0,
			};
		});
		expect(bounds).toEqual({ bottom: true, top: true });

		// AND: scrolling it reaches every control.
		const reachable = await page.evaluate(() => {
			const content = document.querySelector('[role=dialog]');
			if (!content) throw new Error('no dialog');
			content.scrollTop = content.scrollHeight;
			return [...content.querySelectorAll('button')].every(
				(button) =>
					button.getBoundingClientRect().bottom <= globalThis.innerHeight,
			);
		});
		expect(reachable).toBe(true);

		// Dismiss without saving: these specs share the auth instance.
		await page.keyboard.press('Escape');
		await expect(dialog()).toBeHidden();
	});
});
