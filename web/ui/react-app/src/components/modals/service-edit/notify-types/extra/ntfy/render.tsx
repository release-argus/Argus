import { FC, memo } from "react";
import { NotifyNtfyAction, NotifyNtfyActionTypes } from "types/config";

import BROADCAST from "./broadcast";
import HTTP from "./http";
import VIEW from "./view";

interface RenderTypeProps {
  name: string;
  targetType: string;
  defaults?: NotifyNtfyAction;
}

const RENDER_TYPE_COMPONENTS: {
  [key in NotifyNtfyActionTypes]: FC<{
    name: string;
    defaults?: NotifyNtfyAction;
  }>;
} = {
  broadcast: BROADCAST,
  http: HTTP,
  view: VIEW,
};

/**
 *
 * @param name - The name of the field in the form
 * @param targetType - The type of the field
 * @param defaults - The default values for the field
 * @returns The form fields for ntfy.params.actions
 */
const RenderAction: FC<RenderTypeProps> = ({ name, targetType, defaults }) => {
  const RenderTypeComponent =
    RENDER_TYPE_COMPONENTS[(targetType || "view") as NotifyNtfyActionTypes];

  return <RenderTypeComponent name={name} defaults={defaults} />;
};

export default memo(RenderAction);
