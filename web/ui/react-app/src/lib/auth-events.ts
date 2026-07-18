/**
 * Event fired (e.g. by the query error handler) when any API call gets a 401,
 * so the app can drop to the login page.
 */
export const UNAUTHORISED_EVENT = 'argus:unauthorised';

/** Notifies the auth context that a request was rejected as unauthorised. */
export const notifyUnauthorised = () => {
	globalThis.dispatchEvent(new Event(UNAUTHORISED_EVENT));
};
