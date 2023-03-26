import { FC, memo } from "react";
import { REGEX, REPLACE, SPLIT } from ".";

import { URLCommandTypes } from "types/config";

const RENDER_TYPE_COMPONENTS: {
  [key in URLCommandTypes]: FC<{
    name: string;
  }>;
} = {
  regex: REGEX,
  replace: REPLACE,
  split: SPLIT,
};

const RenderURLCommand = ({
  name,
  commandType,
}: {
  name: string;
  commandType: URLCommandTypes;
}) => {
  const RenderTypeComponent = RENDER_TYPE_COMPONENTS[commandType];

  return <RenderTypeComponent name={name} />;
};

export default memo(RenderURLCommand);
