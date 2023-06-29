import {
  BARK,
  DISCORD,
  GOOGLE_CHAT,
  GOTIFY,
  IFTTT,
  JOIN,
  MATRIX,
  MATTERMOST,
  NTFY,
  OPSGENIE,
  PUSHBULLET,
  PUSHOVER,
  ROCKET_CHAT,
  SLACK,
  SMTP,
  TEAMS,
  TELEGRAM,
  ZULIP,
} from "components/modals/service-edit/notify-types";
import { Dict, NotifyType, NotifyTypes } from "types/config";
import { FC, memo } from "react";

interface RenderTypeProps {
  name: string;
  type: string;
  globalNotify?: NotifyType;
  defaults?: Dict<NotifyType>;
  hard_defaults?: NotifyType;
}

const RENDER_TYPE_COMPONENTS: {
  [key in NotifyTypes]: FC<{
    name: string;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    global: any;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    defaults: any;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    hard_defaults: any;
  }>;
} = {
  bark: BARK,
  discord: DISCORD,
  smtp: SMTP,
  googlechat: GOOGLE_CHAT,
  gotify: GOTIFY,
  ifttt: IFTTT,
  join: JOIN,
  mattermost: MATTERMOST,
  matrix: MATRIX,
  ntfy: NTFY,
  opsgenie: OPSGENIE,
  pushbullet: PUSHBULLET,
  pushover: PUSHOVER,
  rocketchat: ROCKET_CHAT,
  slack: SLACK,
  teams: TEAMS,
  telegram: TELEGRAM,
  zulip: ZULIP,
};

const RenderNotify: FC<RenderTypeProps> = ({
  name,
  type,
  globalNotify,
  defaults,
  hard_defaults,
}) => {
  const RenderTypeComponent =
    RENDER_TYPE_COMPONENTS[(type as NotifyTypes) || "discord"];
  return (
    <RenderTypeComponent
      name={name}
      global={globalNotify}
      defaults={defaults}
      hard_defaults={hard_defaults}
    />
  );
};

export default memo(RenderNotify);
