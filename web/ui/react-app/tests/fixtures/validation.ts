import { expect, type Locator, type Page } from '@playwright/test';

// Validation messages for the modal forms.
export const REQUIRED = 'Required.';
export const NUMBER_REQUIRED = 'Number required.';
export const MUST_BE_UNIQUE = 'Must be unique.';

/**
 * The field container for `input` - scopes assertions to one field's error so
 * they don't match other fields' errors.
 *
 * @param input - The field's input element.
 * @returns The enclosing `[data-slot="field"]` container.
 */
export const fieldOf = (input: Locator) =>
	input.locator('xpath=ancestor::*[@data-slot="field"][1]');

/**
 * Opens the "Create a service" modal, entering edit mode first.
 *
 * @param page - The dashboard page.
 * @returns The open modal dialog.
 */
export const openCreateServiceModal = async (page: Page) => {
	await page.goto('/');
	// Wait for /api/v1/service/order response.
	await expect(page.locator('[data-service-id]').first()).toBeVisible();

	await page.getByRole('button', { name: /toggle edit mode/i }).click();
	await page.getByRole('button', { name: /create a service/i }).click();
	const dialog = page.getByRole('dialog');
	await expect(dialog).toBeVisible();
	return dialog;
};

/**
 * Expands the top-level accordion section with the given name.
 *
 * @param dialog - The modal dialog.
 * @param name - The section's heading text (e.g. 'Latest Version').
 * @returns The section's accordion-item container.
 */
export const openSection = async (dialog: Locator, name: string) => {
	await dialog
		.getByRole('button', { name: new RegExp(`^${name}:?$`, 'i') })
		.click();
	return dialog
		.locator('[data-slot="accordion-item"]', { hasText: name })
		.first();
};

/**
 * Waits for an accordion item's resize animations to finish before a
 * screenshot, so error text isn't captured mid-resize.
 *
 * @param section - The accordion item.
 */
export const waitForAccordionAnimations = (section: Locator) =>
	section.evaluate((el) =>
		Promise.all(
			el
				.getAnimations({ subtree: true })
				.map((a) => a.finished.catch(() => undefined)),
		),
	);

/**
 * Activates a button via keyboard (focus + Enter). Used for icon-only
 * add/remove buttons whose empty-array row can intercept the pointer and make
 * `.click()` time out - keyboard activation skips hit-testing.
 *
 * @param button - The button to activate.
 */
export const clickViaKeyboard = async (button: Locator) => {
	await button.focus();
	await button.press('Enter');
};

// A bound screenshot taker (from `screenshotsUnder`): a leaf file name and an
// optional field to centre on.
type Shot = (file: string, centerOn?: Locator) => Promise<Buffer>;

/**
 * Adds a notifier to the (already-open) Notify `section` and expands it.
 *
 * @param section - The Notify accordion section.
 * @param dialog - The modal dialog (for the type-select options).
 * @param type - Visible option label (e.g. 'Gotify') to switch from the default
 *   notifier; omit to leave it as-is.
 */
export const addNotify = async (
	section: Locator,
	dialog: Locator,
	type?: string,
) => {
	await section.getByRole('button', { name: /add notify/i }).click();
	const header = section.locator('[data-slot="accordion-trigger"]', {
		hasText: /^0:/,
	});
	await expect(header).toBeVisible();
	await header.click();
	if (!type) return;
	await section.locator('#notify\\.0\\.type').click();
	await dialog.getByRole('option', { exact: true, name: type }).click();
};

/**
 * The "Value field for <label>" textbox within `section` (the input itself,
 * not its `fieldOf` container).
 *
 * @param section - The accordion section to search within.
 * @param label - The field label, e.g. 'Host'.
 * @returns The textbox locator.
 */
export const valueInputFor = (section: Locator, label: string) =>
	section.getByRole('textbox', { name: `Value field for ${label}` });

/**
 * Fills `input`, blurs it, asserts the field is invalid with `error`, then
 * screenshots it as `name`.
 *
 * @param input - The field input.
 * @param value - Value to fill; `undefined` validates the field without a fill.
 * @param error - The error text/pattern the field must show.
 * @param shot - The screenshot taker.
 * @param name - The screenshot's leaf file name.
 * @param section - Enclosing accordion item to settle its resize animation
 *   before the shot (omit for top-level fields).
 */
export const expectError = async (
	input: Locator,
	value: string | undefined,
	error: string | RegExp,
	shot: Shot,
	name: string,
	section?: Locator,
) => {
	const field = fieldOf(input);
	await input.focus();
	if (value !== undefined) await input.fill(value);
	await input.blur();
	await expect(input).toHaveAttribute('aria-invalid', 'true');
	await expect(field.getByRole('alert')).toHaveText(error);
	if (section) await waitForAccordionAnimations(section);
	await shot(name, field);
};

/**
 * Fills `input`, blurs it, asserts the field has no error, then screenshots it
 * as `name`.
 *
 * @param input - The field input.
 * @param value - Value to fill; `undefined` validates the field without a fill.
 * @param shot - The screenshot taker.
 * @param name - The screenshot's leaf file name.
 * @param section - Enclosing accordion item to settle its resize animation
 *   before the shot (omit for top-level fields).
 */
export const expectValid = async (
	input: Locator,
	value: string | undefined,
	shot: Shot,
	name: string,
	section?: Locator,
) => {
	const field = fieldOf(input);
	await input.focus();
	if (value !== undefined) await input.fill(value);
	await input.blur();
	await expect(input).toHaveAttribute('aria-invalid', 'false');
	await expect(field.getByRole('alert')).not.toBeVisible();
	if (section) await waitForAccordionAnimations(section);
	await shot(name, field);
};

/**
 * For fields that aren't actually validated (e.g. a non-empty default
 * fallback): asserts no error appears. Unlike `expectValid` it doesn't assert
 * `aria-invalid="false"` - a never-invalid field omits the attribute entirely.
 *
 * @param input - The field input.
 * @param value - Value to fill; `undefined` validates the field without a fill.
 * @param shot - The screenshot taker.
 * @param name - The screenshot's leaf file name.
 * @param section - Enclosing accordion item to settle its resize animation
 *   before the shot (omit for top-level fields).
 */
export const expectNoError = async (
	input: Locator,
	value: string | undefined,
	shot: Shot,
	name: string,
	section?: Locator,
) => {
	const field = fieldOf(input);
	await input.focus();
	if (value !== undefined) await input.fill(value);
	await input.blur();
	await expect(input).not.toHaveAttribute('aria-invalid', 'true');
	await expect(field.getByRole('alert')).not.toBeVisible();
	if (section) await waitForAccordionAnimations(section);
	await shot(name, field);
};

// ARIA role for resolving a string `input` by accessible name: `textbox` for
// free-text fields not using the "Value field for <x>" name, `combobox` for
// selects.
export type InputRole = 'textbox' | 'combobox';

// One field's validation case for `runValidations`: enter `bad` and assert it
// shows `error` (or `null` to assert no error), then optionally enter `good`
// and assert it clears. `input` is a locator, or a label string resolved via
// `inputType`/`valueInputFor`.
export type FieldCheck = {
	input: string | Locator;
	inputType?: InputRole;
	bad: string | undefined;
	error: string | RegExp | null;
	badSlug: string;
	good?: string;
	goodSlug?: string;
};

/**
 * Runs each check as [bad -> assert] then optional [good -> assert valid],
 * auto-numbering the screenshots.
 *
 * @param section - The enclosing accordion item (for animation settling).
 * @param shot - The screenshot taker.
 * @param checks - The field cases to run, in order.
 */
export const runValidations = async (
	section: Locator,
	shot: Shot,
	checks: FieldCheck[],
) => {
	let n = 0;
	const idx = () => String(++n).padStart(2, '0');
	for (const c of checks) {
		const input =
			typeof c.input !== 'string'
				? c.input
				: c.inputType
					? section.getByRole(c.inputType, { name: c.input })
					: valueInputFor(section, c.input);
		if (c.error === null)
			await expectNoError(input, c.bad, shot, `${idx()}-${c.badSlug}`, section);
		else
			await expectError(
				input,
				c.bad,
				c.error,
				shot,
				`${idx()}-${c.badSlug}`,
				section,
			);
		if (c.good !== undefined)
			await expectValid(input, c.good, shot, `${idx()}-${c.goodSlug}`, section);
	}
};

// Shared args for the `required`/`numeric`/`optional` constructors below.
type CheckArgs = {
	input: string | Locator;
	inputType?: InputRole;
	good: string;
	slug: string;
};

/**
 * A required-field check: empty shows `REQUIRED`, `good` clears it.
 *
 * @param input - The field label or locator.
 * @param inputType - How to resolve a string `input` (see `FieldCheck`).
 * @param good - The value that must validate.
 * @param slug - The screenshot name stem.
 * @returns The field check.
 */
export const required = ({
	input,
	inputType,
	good,
	slug,
}: CheckArgs): FieldCheck => ({
	bad: undefined,
	badSlug: `${slug}-empty`,
	error: REQUIRED,
	good,
	goodSlug: `${slug}-valid`,
	input,
	inputType,
});

/**
 * A numeric-field check: a non-number shows `NUMBER_REQUIRED`, `good` clears it.
 *
 * @param input - The field label or locator.
 * @param inputType - How to resolve a string `input` (see `FieldCheck`).
 * @param good - The value that must validate.
 * @param slug - The screenshot name stem.
 * @returns The field check.
 */
export const numeric = ({
	input,
	inputType,
	good,
	slug,
}: CheckArgs): FieldCheck => ({
	bad: 'abc',
	badSlug: `${slug}-not-a-number`,
	error: NUMBER_REQUIRED,
	good,
	goodSlug: `${slug}-valid`,
	input,
	inputType,
});

/**
 * A check for a field that isn't validated: empty shows no error, `good` is
 * still accepted.
 *
 * @param input - The field label or locator.
 * @param inputType - How to resolve a string `input` (see `FieldCheck`).
 * @param good - The value that must validate.
 * @param slug - The screenshot name stem.
 * @returns The field check.
 */
export const optional = ({
	input,
	inputType,
	good,
	slug,
}: CheckArgs): FieldCheck => ({
	bad: undefined,
	badSlug: `${slug}-empty-no-error`,
	error: null,
	good,
	goodSlug: `${slug}-valid`,
	input,
	inputType,
});
