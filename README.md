# pallas-proxy

A Cloudflare Worker that proxies Apple's `gdmf` software-update endpoints and
transparently decodes their JWS-signed responses into plain JSON.

Apple's `gdmf.apple.com` serves its update metadata (e.g. `/v2/pmv`) as a
[JWS](https://datatracker.ietf.org/doc/html/rfc7515) token of the form
`header.payload.signature`, which is awkward to read or pipe into other tools.
This proxy returns the decoded `payload` as JSON instead, while leaving every
other response untouched.

## Routing

The upstream is chosen by the incoming hostname:

| Incoming host             | Upstream                       |
| ------------------------- | ------------------------------ |
| `gdmf.notzheng.com`       | `https://gdmf.apple.com`       |
| `gdmf-ados.notzheng.com`  | `https://gdmf-ados.apple.com`  |

Any other host (e.g. the `*.workers.dev` preview) falls back to
`gdmf.apple.com`.

## Behavior

- Forwards every request (path, query string, method, body) to the matched
  upstream.
- Leaves non-`application/json` responses untouched.
- For JSON responses that look like a JWS token (body starts with the base64
  header prefix `e`), the `payload` segment is base64url-decoded and returned as
  the response body, with `content-encoding` dropped and `content-length`
  recalculated.

## Usage

```sh
curl https://gdmf.notzheng.com/v2/pmv | jq .
```

## Develop

```sh
npm install
npm run typecheck
```

Entry point is `src/index.ts`.

## Deploy

Deployment is wired up through the Cloudflare dashboard's Git integration, so no
`wrangler` config lives in this repo. To serve the custom hostnames, add
`gdmf.notzheng.com` and `gdmf-ados.notzheng.com` as Custom Domains on the Worker
(the `notzheng.com` zone must be on Cloudflare DNS).

> TLS certificate verification is intentionally not disabled: the Worker fetches
> the real Apple hostnames directly, so SNI matches their valid certificates.
> Cloudflare's runtime also offers no way to skip verification.
