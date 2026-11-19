// biome-ignore assist/source/useSortedKeys: ascending order.
export const MQTT_QOS = {
	AT_MOST_ONCE: { label: '0 - At most once', value: 'AtMostOnce' },
	AT_LEAST_ONCE: { label: '1 - At least once', value: 'AtLeastOnce' },
	EXACTLY_ONCE: { label: '2 - Exactly once', value: 'ExactlyOnce' },
} as const;
export type MQTTQoS = (typeof MQTT_QOS)[keyof typeof MQTT_QOS]['value'];
export const mqttQoSOptions = Object.values(MQTT_QOS);
