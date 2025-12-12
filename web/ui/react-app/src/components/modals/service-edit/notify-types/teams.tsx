import { useMemo } from 'react';
import { FieldColour, FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyTeamsSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Teams` notifier.
 *
 * @param name - The path to this `Teams` in the form.
 * @param main - The main values.
 */
const TEAMS = ({ name, main }: { name: string; main?: NotifyTeamsSchema }) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.teams,
		[main, typeDataDefaults?.notify.teams],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					defaultVal={defaults?.url_fields?.group}
					label="Group"
					name={`${name}.url_fields.group`}
					required
					tooltip={{
						ariaLabel:
							'Format: https://<organisation>.webhook.office.com/webhookb2/<GROUP>@<tenant>/IncomingWebhook/<altID>/<groupOwner>/<extraId>',
						content: (
							<>
								<span className="text-muted-foreground">
									{'e.g. https://<organisation>.webhook.office.com/webhookb2/'}
								</span>
								<span className="bold underline">{'<GROUP>'}</span>
								<span className="text-muted-foreground">
									{'@<tenant>/IncomingWebhook/<altID>/<groupOwner>/<extraId>'}
								</span>
							</>
						),
						type: 'element',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.tenant}
					label="Tenant"
					name={`${name}.url_fields.tenant`}
					required
					tooltip={{
						ariaLabel:
							'Format: https://<organisation>.webhook.office.com/webhookb2/<group>@<TENANT>/IncomingWebhook/<altID>/<groupOwner>/<extraId>',
						content: (
							<>
								<span className="text-muted-foreground">
									{
										'e.g. https://<organisation>.webhook.office.com/webhookb2/<group>@'
									}
								</span>
								<span className="bold underline">{'<TENANT>'}</span>
								<span className="text-muted-foreground">
									{'/IncomingWebhook/<altID>/<groupOwner>/<extraId>'}
								</span>
							</>
						),
						type: 'element',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.altid}
					label="Alt ID"
					name={`${name}.url_fields.altid`}
					required
					tooltip={{
						ariaLabel:
							'Format: https://<organisation>.webhook.office.com/webhookb2/<group>@<tenant>/IncomingWebhook/<ALTID>/<groupOwner>/<extraId>',
						content: (
							<>
								<span className="text-muted-foreground">
									{
										'e.g. https://<organisation>.webhook.office.com/webhookb2/<group>@<tenant>/IncomingWebhook/'
									}
								</span>
								<span className="bold underline">{'<ALTID>'}</span>
								<span className="text-muted-foreground">
									{'/<groupOwner>/<extraId>'}
								</span>
							</>
						),
						type: 'element',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.groupowner}
					label="Group Owner"
					name={`${name}.url_fields.groupowner`}
					required
					tooltip={{
						ariaLabel:
							'Format: https://<organisation>.webhook.office.com/webhookb2/<group>@<tenant>/IncomingWebhook/<altID>/<GROUPONWER>/<extraId>',
						content: (
							<>
								<span className="text-muted-foreground">
									{
										'e.g. https://<organisation>.webhook.office.com/webhookb2/<group>@<tenant>/IncomingWebhook/<altID>/'
									}
								</span>
								<span className="bold underline">{'<GROUPOWNER>'}</span>
								<span className="text-muted-foreground">{'/<extraId>'}</span>
							</>
						),
						type: 'element',
					}}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url_fields?.extraid}
					label="Extra ID"
					name={`${name}.url_fields.extraid`}
					required
					tooltip={{
						ariaLabel:
							'Format: https://<organisation>.webhook.office.com/webhookb2/<group>@<tenant>/IncomingWebhook/<altID>/<groupOwner>/<EXTRAID>',
						content: (
							<>
								<span className="text-muted-foreground">
									{
										'e.g. https://<organisation>.webhook.office.com/webhookb2/<group>@<tenant>/IncomingWebhook/<altID>/<groupOwner>/ '
									}
								</span>
								<span className="bold underline">{'<EXTRAID>'}</span>
							</>
						),
						type: 'element',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					colSize={{ xs: 6 }}
					defaultVal={defaults?.params?.host}
					label="Host"
					name={`${name}.params.host`}
					required
					tooltip={{
						content:
							'This is the "<organization>.webhook.office.com" in your webhook',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ xs: 6 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
				<FieldColour
					defaultVal={defaults?.params?.color}
					label="Color"
					name={`${name}.params.color`}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default TEAMS;
