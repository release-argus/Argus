import { Accordion, Container, Form, Stack } from "react-bootstrap";

import { FC } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormLabel } from "components/generic/form";
import { faCircleNotch } from "@fortawesome/free-solid-svg-icons";
import { useDelayedRender } from "hooks/delayed-render";

interface Props {
  name: string;
}
export const Loading: FC<Props> = ({ name }) => {
  const delayedRender = useDelayedRender(500);
  const accordionHeaders = [
    "Options:",
    "Latest Version:",
    "Deployed Version:",
    "Commands:",
    "WebHooks:",
    "Notify:",
    "Dashboard:",
  ];
  const formControlClassName = "pt-1 pb-1 col-form col-sm-12 col-12";
  return (
    <Stack gap={3}>
      <Form.Group className="mb-2">
        <Form.Group className={formControlClassName}>
          <FormLabel text="Name" required />
          <Form.Control
            autoFocus={false}
            defaultValue={name}
            disabled
            className="bg-transparent"
          />
        </Form.Group>
        <Form.Group className={formControlClassName}>
          <FormLabel text="Comment" />
          <Form.Control autoFocus={false} disabled className="bg-transparent" />
        </Form.Group>
        {delayedRender(() => (
          <Container className="empty">
            <FontAwesomeIcon icon={faCircleNotch} className={"fa-spin"} />
            <span style={{ paddingLeft: "0.5rem" }}>Loading...</span>
          </Container>
        ))}
      </Form.Group>
      {accordionHeaders.map((title) => {
        return (
          <Accordion key={title}>
            <Accordion.Header>{title}</Accordion.Header>
          </Accordion>
        );
      })}
    </Stack>
  );
};
