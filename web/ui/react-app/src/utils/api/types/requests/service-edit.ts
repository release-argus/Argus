export type ServiceEditRequestBuilder = {
	/* The service ID. */
	serviceID: string | null;
	/* Service JSON */
	body: unknown;
};

export type ServiceEditResponse = {
	/* The result of the edit. */
	message: string;
};
