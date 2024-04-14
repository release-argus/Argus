import { Accordion, Button, Stack } from "react-bootstrap";
import { Dict, NotifyType } from "types/config";
import { FC, memo, useCallback, useMemo } from "react";

import Notify from "./notify";
import { NotifyEditType } from "types/service-edit";
import { isEmptyArray } from "utils";
import { useFieldArray } from "react-hook-form";

interface Props {
  serviceName: string;

  originals?: NotifyEditType[];
  mains?: Dict<NotifyType>;
  defaults?: Dict<NotifyType>;
  hard_defaults?: Dict<NotifyType>;
}

/**
 * Returns the form fields for `notify`
 *
 * @param serviceName - The name of the service
 * @param originals - The original values in the form
 * @param mains - The main notify's
 * @param defaults - The default values for each `notify` types
 * @param hard_defaults - The hard default values for each `notify` types
 * @returns The form fields for `notify`
 */
const EditServiceNotifys: FC<Props> = ({
  serviceName,

  originals,
  mains,
  defaults,
  hard_defaults,
}) => {
  const { fields, append, remove } = useFieldArray({
    name: "notify",
  });
  const addItem = useCallback(() => {
    append(
      {
        type: "discord",
        name: "",
        options: {},
        url_fields: {},
        params: { avatar: "", color: "", icon: "" },
      },
      { shouldFocus: false }
    );
  }, []);

  const globalNotifyOptions = useMemo(
    () => (
      <>
        <option className="form-select-option" value="">
          --Not global--
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
      <Accordion.Header>Notify:</Accordion.Header>
      <Accordion.Body>
        <Stack gap={2}>
          {fields.map(({ id }, index) => (
            <Notify
              key={id}
              name={`notify.${index}`}
              removeMe={() => remove(index)}
              serviceName={serviceName}
              originals={originals}
              globalOptions={globalNotifyOptions}
              mains={mains}
              defaults={defaults}
              hard_defaults={hard_defaults}
            />
          ))}
          <Button
            className={isEmptyArray(fields) ? "mt-2" : ""}
            variant="secondary"
            style={{ width: "100%", marginTop: "1rem" }}
            onClick={addItem}
          >
            Add Notify
          </Button>
        </Stack>
      </Accordion.Body>
    </Accordion>
  );
};

export default memo(EditServiceNotifys);
