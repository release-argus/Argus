import { Accordion, Button, Stack } from "react-bootstrap";
import { Dict, WebHookType } from "types/config";
import { FC, useCallback, useMemo } from "react";

import EditServiceWebHook from "components/modals/service-edit/webhook";
import { firstNonEmpty } from "utils";
import { useFieldArray } from "react-hook-form";

interface Props {
  mains?: Dict<WebHookType>;
  defaults?: WebHookType;
  hard_defaults?: WebHookType;
}

/**
 * Returns the form fields for `webhook`
 *
 * @param mains - The global WebHook's
 * @param defaults - The default values for a WebHook
 * @param hard_defaults - The hard default values for a WebHook
 * @returns The form fields for `webhook`
 */
const EditServiceWebHooks: FC<Props> = ({ mains, defaults, hard_defaults }) => {
  const { fields, append, remove } = useFieldArray({
    name: "webhook",
  });
  const convertedDefaults = useMemo(
    () => ({
      custom_headers: firstNonEmpty(
        defaults?.custom_headers,
        hard_defaults?.custom_headers
      ).map(() => ({ key: "", item: "" })),
    }),
    [defaults, hard_defaults]
  );
  const addItem = useCallback(() => {
    append(
      {
        type: "github",
        name: "",
        custom_headers: convertedDefaults.custom_headers,
      },
      { shouldFocus: false }
    );
  }, []);

  const globalWebHookOptions = useMemo(
    () => (
      <>
        <option className="form-select-option" value="">
          Not global
        </option>
        {mains &&
          Object.keys(mains).map((n) => (
            <option className="form-select-option" value={n} key={n}>
              {n}
            </option>
          ))}
      </>
    ),
    [mains]
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
              mains={mains}
              defaults={defaults}
              hard_defaults={hard_defaults}
            />
          ))}
          <Button
            className={fields.length > 0 ? "" : "mt-2"}
            variant="secondary"
            style={{ width: "100%", marginTop: "1rem" }}
            onClick={addItem}
          >
            Add WebHook
          </Button>
        </Stack>
      </Accordion.Body>
    </Accordion>
  );
};

export default EditServiceWebHooks;
