import {
  FormItem,
  FormItemWithPreview,
  FormLabel,
  FormSelect,
} from "components/generic/form";
import { useEffect, useMemo } from "react";

import { NotifyBarkType } from "types/config";
import NotifyOptions from "components/modals/service-edit/notify-types/shared";
import { globalOrDefault } from "components/modals/service-edit/util";
import { normaliseForSelect } from "components/modals/service-edit/util";
import { useFormContext } from "react-hook-form";

export const BarkSchemeOptions = [
  { label: "HTTPS", value: "https" },
  { label: "HTTP", value: "http" },
];

export const BarkSoundOptions = [
  { label: "", value: "" },
  { label: "Alarm", value: "alarm" },
  { label: "Anticipate", value: "anticipate" },
  { label: "Bell", value: "bell" },
  { label: "Birdsong", value: "birdsong" },
  { label: "Bloom", value: "bloom" },
  { label: "Calypso", value: "calypso" },
  { label: "Chime", value: "chime" },
  { label: "Choo", value: "choo" },
  { label: "Descent", value: "descent" },
  { label: "Electronic", value: "electronic" },
  { label: "Fanfare", value: "fanfare" },
  { label: "Glass", value: "glass" },
  { label: "GoToSleep", value: "gotosleep" },
  { label: "HealthNotification", value: "healthnotification" },
  { label: "Horn", value: "horn" },
  { label: "Ladder", value: "ladder" },
  { label: "MailSent", value: "maisentl" },
  { label: "Minuet", value: "minuet" },
  { label: "MultiWayInvitation", value: "multiwayinvitation" },
  { label: "NewMail", value: "newmail" },
  { label: "NewsFlash", value: "newsflash" },
  { label: "Noir", value: "noir" },
  { label: "PaymentSuccess", value: "paymentsuccess" },
  { label: "Shake", value: "shake" },
  { label: "SherwoodForest", value: "sherwoodforest" },
  { label: "Silence", value: "silence" },
  { label: "Spell", value: "spell" },
  { label: "Suspense", value: "suspense" },
  { label: "Telegraph", value: "telegraph" },
  { label: "Tiptoes", value: "tiptoes" },
  { label: "Typewriters", value: "typewriters" },
  { label: "Update", value: "update" },
];

/**
 * Returns the form fields for `Bark`
 *
 * @param name - The path to this `Bark` in the form
 * @param main - The main values
 * @param defaults - The default values
 * @param hard_defaults - The hard default values
 * @returns The form fields for this `Bark` `Notify`
 */
const BARK = ({
  name,

  main,
  defaults,
  hard_defaults,
}: {
  name: string;

  main?: NotifyBarkType;
  defaults?: NotifyBarkType;
  hard_defaults?: NotifyBarkType;
}) => {
  const { getValues, setValue } = useFormContext();

  const defaultParamsScheme = globalOrDefault(
    main?.params?.scheme,
    defaults?.params?.scheme,
    hard_defaults?.params?.scheme
  ).toLowerCase();
  const barkSchemeOptions = useMemo(() => {
    const defaultScheme = normaliseForSelect(
      BarkSchemeOptions,
      defaultParamsScheme
    );

    if (defaultScheme)
      return [
        { value: "", label: `${defaultScheme.label} (default)` },
        ...BarkSchemeOptions,
      ];

    return BarkSchemeOptions;
  }, [defaultParamsScheme]);

  const defaultParamsSound = globalOrDefault(
    main?.params?.sound,
    defaults?.params?.sound,
    hard_defaults?.params?.sound
  ).toLowerCase();
  const barkSoundOptions = useMemo(() => {
    const defaultSound = normaliseForSelect(
      BarkSoundOptions,
      defaultParamsSound
    );

    if (defaultSound)
      return [
        { value: "", label: `${defaultSound.label} (default)` },
        ...BarkSoundOptions.filter((option) => option.value !== ""),
      ];

    return BarkSoundOptions;
  }, [defaultParamsSound]);

  useEffect(() => {
    // Normalise selected scheme, or default it
    if (defaultParamsScheme === "")
      setValue(
        `${name}.params.scheme`,
        normaliseForSelect(
          BarkSchemeOptions,
          getValues(`${name}.params.scheme`)
        )?.value || "https"
      );

    // Normalise selected sound, or default it
    if (
      defaultParamsSound === "" &&
      getValues(`${name}.params.sound`) !== undefined
    )
      setValue(
        `${name}.params.sound`,
        normaliseForSelect(BarkSoundOptions, getValues(`${name}.params.sound`))
          ?.value || ""
      );
  }, []);

  return (
    <>
      <NotifyOptions
        name={name}
        main={main?.options}
        defaults={defaults?.options}
        hard_defaults={hard_defaults?.options}
      />
      <>
        <FormLabel text="URL Fields" heading />
        <FormItem
          name={`${name}.url_fields.host`}
          col_sm={9}
          required
          label="Host"
          defaultVal={globalOrDefault(
            main?.url_fields?.host,
            defaults?.url_fields?.host,
            hard_defaults?.url_fields?.host
          )}
          onRight
        />
        <FormItem
          name={`${name}.url_fields.port`}
          col_sm={3}
          required
          label="Port"
          defaultVal={globalOrDefault(
            main?.url_fields?.port,
            defaults?.url_fields?.port,
            hard_defaults?.url_fields?.port
          )}
        />
        <FormItem
          name={`${name}.url_fields.path`}
          label="Path"
          tooltip="Server path"
          defaultVal={globalOrDefault(
            main?.url_fields?.path,
            defaults?.url_fields?.path,
            hard_defaults?.url_fields?.path
          )}
        />
        <FormItem
          name={`${name}.url_fields.devicekey`}
          required
          label="Device Key"
          defaultVal={globalOrDefault(
            main?.url_fields?.devicekey,
            defaults?.url_fields?.devicekey,
            hard_defaults?.url_fields?.devicekey
          )}
          onRight
        />
      </>
      <>
        <FormLabel text="Params" heading />
        <FormSelect
          name={`${name}.params.scheme`}
          col_sm={3}
          label="Scheme"
          tooltip="Server protocol"
          options={barkSchemeOptions}
        />
        <FormItem
          name={`${name}.params.badge`}
          col_sm={3}
          type="number"
          label="Badge"
          tooltip="The number displayed next to the App icon"
          defaultVal={globalOrDefault(
            main?.params?.badge,
            defaults?.params?.badge,
            hard_defaults?.params?.badge
          )}
        />
        <FormItem
          name={`${name}.params.copy`}
          label="Copy"
          tooltip="The value to be copied"
          defaultVal={globalOrDefault(
            main?.params?.copy,
            defaults?.params?.copy,
            hard_defaults?.params?.copy
          )}
        />
        <FormItem
          name={`${name}.params.group`}
          label="Group"
          tooltip="The group of the notification"
          defaultVal={globalOrDefault(
            main?.params?.group,
            defaults?.params?.group,
            hard_defaults?.params?.group
          )}
        />
        <FormSelect
          name={`${name}.params.sound`}
          col_sm={6}
          label="Sound"
          options={barkSoundOptions}
        />
        <FormItem
          name={`${name}.params.title`}
          label="Title"
          defaultVal={globalOrDefault(
            main?.params?.title,
            defaults?.params?.title,
            hard_defaults?.params?.title
          )}
        />
        <FormItem
          name={`${name}.params.url`}
          label="URL"
          tooltip="URL to open when notification is tapped"
          defaultVal={globalOrDefault(
            main?.params?.url,
            defaults?.params?.url,
            hard_defaults?.params?.url
          )}
        />
        <FormItemWithPreview
          name={`${name}.params.icon`}
          label="Icon"
          tooltip="URL to an icon"
          defaultVal={globalOrDefault(
            main?.params?.icon,
            defaults?.params?.icon,
            hard_defaults?.params?.icon
          )}
        />
      </>
    </>
  );
};

export default BARK;
