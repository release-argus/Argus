import { BooleanWithDefault } from '@/components/generic';
import { FieldCheck, FieldText } from '@/components/generic/field';
import {
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from '@/components/ui/accordion';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';

/**
 * The `service.options` form fields.
 *
 * @returns The form fields for the `options`.
 */
const EditServiceOptions = () => {
	const name = 'options';
	const { schemaDataDefaults } = useSchemaContext();

	return (
		<AccordionItem value="options">
			<AccordionTrigger>Options:</AccordionTrigger>
			<AccordionContent className="grid grid-cols-12 gap-2">
				<FieldCheck
					checkboxClassName="max-w-9"
					colSize={{ sm: 4, xs: 4 }}
					label="Active"
					name={`${name}.active`}
					tooltip={{
						content: 'Whether the service is active and checking for updates',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 8, xs: 8 }}
					defaultVal={schemaDataDefaults?.options?.interval}
					key="interval"
					label="Interval"
					name={`${name}.interval`}
					tooltip={{
						content:
							'How often to check for both latest version and deployed version updates',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={schemaDataDefaults?.options?.semantic_versioning}
					label="Semantic versioning"
					name={`${name}.semantic_versioning`}
					tooltip={{
						content: "Releases follow 'MAJOR.MINOR.PATCH' versioning",
						type: 'string',
					}}
				/>
			</AccordionContent>
		</AccordionItem>
	);
};

export default EditServiceOptions;
