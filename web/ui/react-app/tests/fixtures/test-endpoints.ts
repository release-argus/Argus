// Test endpoints mirrored from `test/secrets.go` - keep in sync with it.

// Domain with a valid TLS certificate.
export const VALID_CERT_NO_PROTOCOL = 'valid.release-argus.io';
export const VALID_CERT_HTTPS = `https://${VALID_CERT_NO_PROTOCOL}`;

// Domain with an invalid/self-signed TLS certificate.
export const INVALID_CERT_HTTPS = 'https://invalid.release-argus.io';

/**
 * Builds a `/bare/<url-encoded body>` URL on the test server, which echoes
 * `body` back as the response. Lets a test pin the exact lookup response.
 *
 * @param body - The exact response body the endpoint should return.
 * @param invalidCert - Target the invalid-cert host instead of the valid one.
 */
export const bareEndpoint = (body: string, invalidCert = false) =>
	`${invalidCert ? INVALID_CERT_HTTPS : VALID_CERT_HTTPS}/bare/${encodeURIComponent(body)}`;

// Version carried in a response header.
export const LOOKUP_RESPONSE_HEADER = {
	headerKeyFail: 'X-Version-Foo',
	headerKeyPass: 'X-Version-Here',
	headerKeyPassMixedCase: 'x-VeRSioN-HERe',
	urlInvalid: `${INVALID_CERT_HTTPS}/header`,
	urlValid: `${VALID_CERT_HTTPS}/header`,
};

// Requires a custom request header to succeed.
export const LOOKUP_WITH_HEADER_AUTH = {
	headerKey: 'X-Test',
	headerValueFail: 'secret-',
	headerValuePass: 'secret',
	urlInvalid: `${INVALID_CERT_HTTPS}/hooks/single-header`,
	urlValid: `${VALID_CERT_HTTPS}/hooks/single-header`,
};

// Requires HTTP basic auth to succeed.
export const LOOKUP_BASIC_AUTH = {
	password: '123',
	urlInvalid: `${INVALID_CERT_HTTPS}/basic-auth`,
	urlValid: `${VALID_CERT_HTTPS}/basic-auth`,
	username: 'test',
};

// WebHook receiver requiring a matching secret.
export const WEBHOOK_GITHUB = {
	secretFail: 'argus-',
	secretPass: 'argus',
	urlInvalid: `${INVALID_CERT_HTTPS}/hooks/github-style`,
	urlValid: `${VALID_CERT_HTTPS}/hooks/github-style`,
};

// A real Gotify endpoint that accepts `tokenPass`, rejecting anything else.
export const NOTIFY_GOTIFY = {
	host: VALID_CERT_NO_PROTOCOL,
	path: '/gotify',
	// trunk-ignore(gitleaks/generic-api-key)
	tokenPass: 'AGE-LlHU89Q56uQ',
};
