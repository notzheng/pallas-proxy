const HOST_MAP: Record<string, string> = {
  "gdmf.notzheng.com": "https://gdmf.apple.com",
  "gdmf-ados.notzheng.com": "https://gdmf-ados.apple.com",
};

const DEFAULT_ORIGIN = "https://gdmf.apple.com";

export default {
  async fetch(request: Request): Promise<Response> {
    const incoming = new URL(request.url);
    const targetOrigin = HOST_MAP[incoming.hostname] ?? DEFAULT_ORIGIN;
    const targetUrl = targetOrigin + incoming.pathname + incoming.search;

    const headers = new Headers(request.headers);
    headers.delete("host");

    const hasBody = request.method !== "GET" && request.method !== "HEAD";
    const upstream = await fetch(targetUrl, {
      method: request.method,
      headers,
      body: hasBody ? request.body : undefined,
      redirect: "manual",
    });

    const mimeType = (upstream.headers.get("content-type") ?? "")
      .split(";")[0]
      .trim()
      .toLowerCase();

    if (mimeType !== "application/json") {
      return upstream;
    }

    const bodyText = await upstream.text();

    // Apple's gdmf responses are JWS-signed; the raw body is `header.payload.signature`.
    // A plain (unsigned) JSON body won't start with the base64 JWT header prefix.
    if (!bodyText.startsWith("e")) {
      return new Response(bodyText, upstream);
    }

    const payload = bodyText.split(".")[1];
    const decoded = base64ToBytes(payload);

    const responseHeaders = new Headers(upstream.headers);
    responseHeaders.delete("content-encoding");
    responseHeaders.set("content-length", String(decoded.byteLength));

    return new Response(decoded, {
      status: upstream.status,
      statusText: upstream.statusText,
      headers: responseHeaders,
    });
  },
} satisfies ExportedHandler;

function base64ToBytes(value: string): Uint8Array {
  let normalized = value.replace(/-/g, "+").replace(/_/g, "/");
  const remainder = normalized.length % 4;
  if (remainder) {
    normalized += "=".repeat(4 - remainder);
  }
  const binary = atob(normalized);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}
