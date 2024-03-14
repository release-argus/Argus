type Props = {
  // The URL to fetch data from
  url: string;
  // The HTTP method to use, either GET or POST
  method?: "GET" | "POST";
  // Optional headers to include in the request
  headers?: Record<string, string>;
  // Optional request body, applicable for POST requests
  body?: string;
};

// fetchJSON will GET the JSON data, rewritten for use in React-Query
const fetchJSON = async <T,>({url, method="GET", headers, body}: Props): Promise<T> => {
  const response = await Promise.race([
    fetch(url,{
      method: method,
      headers: headers,
      body: body
    }),
    new Promise<Response>((_, reject) =>
      setTimeout(() => reject(new Error("Timeout")), 10000)
    ),
  ]);
  const resp = await response.json();
  return resp;
};

export default fetchJSON;
