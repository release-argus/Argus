import { parse } from "yaml";

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
