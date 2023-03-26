import { Accordion, Form, Placeholder, Stack } from "react-bootstrap";

import { FormLabel } from "components/generic/form";
import { useDelayedRender } from "hooks/delayed-render";

export const Loading = () => {
  const delayedRender = useDelayedRender(250);
  const accordionHeaders = [
    "Options:",
    "Latest Version:",
    "Deployed Version:",
    "Commands:",
    "WebHooks:",
    "Notify:",
    "Dashboard:",
  ];
  return (
    <Stack gap={3}>
      <Form.Group className="mb-2">
        <Form.Group className="mb-2">
          <FormLabel text="Name" required />
          {delayedRender(() => (
            <Placeholder xs={12} style={{ fontSize: "2rem" }} />
          ))}
        </Form.Group>
        <Form.Group className="mb-2">
          <FormLabel text="Comment" />
          {delayedRender(() => (
            <Placeholder xs={12} style={{ fontSize: "2rem" }} />
          ))}
        </Form.Group>
      </Form.Group>
      <br />
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
