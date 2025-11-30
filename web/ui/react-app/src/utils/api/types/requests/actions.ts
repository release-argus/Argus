export type ActionGetRequestBuilder = {
	/* The service ID. */
	serviceID: string;
};

export type ActionSendRequestBuilder = {
	/* The service ID. */
	serviceID: string;
	/* The target action to run */
	target: string;
};
