# Authentication

External shipper and carrier integrations send two credentials:

| Credential | Purpose | Where to obtain it |
| --- | --- | --- |
| OAuth client ID and secret | Obtain a bearer token identifying your integration and role scopes | Request sandbox access from support@oway.io |
| Company API key | Select the company and role whose data you may access | Company portal, Settings → API Keys, after API access is enabled |

Use a shipper key (`oway_sk_...`) for shipper operations or a carrier key (`oway_ck_...`) for carrier operations. Credentials must be provisioned for the chosen environment. Do not infer environment permission from a key's prefix alone.

The SDK exchanges client credentials and caches the token. Set `apiKey` in TypeScript or `APIKey` in Go, or supply the supported per-request company-key override. An optional SDK argument means the key can come from configuration; it does not make the API header optional.

For cURL, POST JSON `{ "clientId": "...", "clientSecret": "..." }` to `/v1/auth/token`. Read `accessToken`, `tokenType`, and `expiresIn` from the response. Other requests send `Authorization: Bearer <accessToken>` and `x-oway-api-key: <company key>`. Cache the token and refresh before `expiresIn` seconds elapse. Never store credentials in source or print tokens in diagnostic output.

A 401 requires checking bearer credentials/expiry; a 403 requires checking the role scope, environment and company-key permissions. Do not retry either repeatedly with unchanged credentials. See [environments](ENVIRONMENTS.md) and [API reference](https://docs.oway.io/reference/gettoken).
