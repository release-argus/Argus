import { parse } from "yaml";

/**
 * Returns the YAML response from the server
 *
 * @param url - The URL to fetch the YAML from
 * @returns The parsed YAML data from the server
 */
const fetchYAML = async (url: string) => {
  const response = await Promise.race([
    fetch(url),
    new Promise<Response>((_, reject) =>
      setTimeout(() => reject(new Error("Timeout")), 10000)
    ),
  ]);

  const text = await response.text();
  return parse(text);
};

export default fetchYAML;
