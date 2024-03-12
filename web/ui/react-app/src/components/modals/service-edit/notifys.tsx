import { Accordion, Button, Stack } from "react-bootstrap";
import { Dict, NotifyType } from "types/config";
import { FC, memo, useCallback, useMemo } from "react";

import Notify from "./notify";
import { NotifyEditType } from "types/service-edit";
import { useFieldArray } from "react-hook-form";

interface Props {
  serviceName: string;

  originals?: NotifyEditType[];
  globals?: Dict<NotifyType>;
  defaults?: Dict<NotifyType>;
  hard_defaults?: Dict<NotifyType>;
}

const EditServiceNotifys: FC<Props> = ({
  serviceName,

  originals,
  globals,
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
              globalNotifyOptions={globalNotifyOptions}
              globals={globals}
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
            Add Notify
          </Button>
        </Stack>
      </Accordion.Body>
    </Accordion>
  );
};

export default memo(EditServiceNotifys);
