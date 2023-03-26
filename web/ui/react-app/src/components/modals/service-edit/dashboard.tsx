import { FC, memo } from "react";
import { FormItem, FormItemWithPreview } from "components/generic/form";

import { Accordion } from "react-bootstrap";
import { BooleanWithDefault } from "components/generic";
import { ServiceDashboardOptionsType } from "types/config";

interface Props {
  dashboard?: ServiceDashboardOptionsType;
  defaults?: ServiceDashboardOptionsType;
  hard_defaults?: ServiceDashboardOptionsType;
}

const EditServiceDashboard: FC<Props> = ({
  dashboard,
  defaults,
  hard_defaults,
}) => (
  <Accordion>
    <Accordion.Header>Dashboard:</Accordion.Header>
    <Accordion.Body>
      <BooleanWithDefault
        name={"dashboard.auto_approve"}
        label="Auto-approve"
        tooltip="Send all commands/webhooks when a new release is found"
        value={dashboard?.auto_approve}
        defaultValue={defaults?.auto_approve || hard_defaults?.auto_approve}
      />
      <FormItemWithPreview
        name={"dashboard.icon"}
        label="Icon"
        tooltip="e.g. https://example.com/icon.png"
        placeholder={defaults?.icon || hard_defaults?.icon}
      />
      <FormItem
        key="icon_link_to"
        name={"dashboard.icon_link_to"}
        col_sm={12}
        label="Icon link to"
        tooltip="Where the Icon will redirect when clicked"
        placeholder={defaults?.icon_link_to || hard_defaults?.icon_link_to}
        isURL
      />
      <FormItem
        key="web_url"
        name={"dashboard.web_url"}
        col_sm={12}
        label="Web URL"
        tooltip="Where the 'Service name' will redirect when clicked"
        placeholder={defaults?.web_url || hard_defaults?.web_url}
        isURL
      />
    </Accordion.Body>
  </Accordion>
);

export default memo(EditServiceDashboard);
