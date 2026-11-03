# Environments

| Environment | Base URL |
| --- | --- |
| Sandbox | `https://api.sandbox.oway.io` |
| Production | `https://api.oway.io` |

Set the sandbox base URL explicitly in first-run examples. Obtain environment-scoped credentials and a synthetic scenario from Oway support before running writes. Sandbox requests create persistent sandbox records; do not assume an arbitrary lane is serviceable, an offer exists, or a reset happens automatically.

For a shipper exercise, request an API-enabled test company with confirmation prerequisites satisfied, a serviceable synthetic lane and document availability instructions. For a carrier exercise, request a pending offer assigned to the test carrier and permission to progress its shipment. Ask for a fresh offer for each complete run. Do not reuse a delivered shipment as a new acceptance exercise.

Only operate on the scenario assigned to you. Use a new run reference and create idempotency key for a new shipment; reuse the key and exact body when reconciling that same creation. Do not delete shared records to reset an exercise. Keep sandbox and production keys separate and explicitly review configuration before moving to production.

See [authentication](AUTH_GUIDE.md) and [quickstarts](examples/quickstarts/README.md).
