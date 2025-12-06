export type TemplateParseRequestBuilder = {
	/* The service ID. */
	serviceID: string;
	/* The template to parse. */
	template: string;
	/* Extra parameters to pass to the template. */
	extraParams?: Record<string, unknown>;
};
export type TemplateParseRequest = {
	service_id: string;
	template: string;
	params?: string;
};

export type TemplateParseResponse = {
	/* The parsed template. */
	parsed: string;
};
