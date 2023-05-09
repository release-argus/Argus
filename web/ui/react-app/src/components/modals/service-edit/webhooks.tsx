import { Accordion, Button, Stack } from "react-bootstrap";
import { FC, useMemo } from "react";
import { ServiceDict, WebHookType } from "types/config";
import { useFieldArray } from "react-hook-form";

import EditServiceWebHook from "components/modals/service-edit/webhook";

interface Props {
  globals?: ServiceDict<WebHookType>;
  defaults?: WebHookType;
  hard_defaults?: WebHookType;
}

const EditServiceWebHooks: FC<Props> = ({
  globals,
  defaults,
  hard_defaults,
}) => {
  const { fields, append, remove } = useFieldArray({
    name: "webhook",
  });

  const globalWebHookOptions = useMemo(
    () => (
      <>
        <option className="form-select-option" value="">
          Not global
        </option>
        {globals &&
          Object.keys(globals).map((n) => (
            <option className="form-select-option" value={n} key={n}>
              {n}
            </option>
          ))}
      </>
    ),
    [globals]
  );

  return (
    <Accordion>
      <Accordion.Header>WebHook:</Accordion.Header>
      <Accordion.Body>
        <Stack gap={2}>
          {fields.map(({ id }, index) => (
            <EditServiceWebHook
              key={id}
              name={`webhook.${index}`}
              removeMe={() => remove(index)}
              globalOptions={globalWebHookOptions}
              globals={globals}
              defaults={defaults}
              hard_defaults={hard_defaults}
            />
          ))}
          <Button
            className={fields.length > 0 ? "" : "mt-2"}
            variant="secondary"
            style={{ width: "100%", marginTop: "1rem" }}
            onClick={() => {
              append({ type: "github", name: "" }, { shouldFocus: false });
            }}
          >
            Add WebHook
          </Button>
        </Stack>
      </Accordion.Body>
    </Accordion>
  );
};

export default EditServiceWebHooks;
