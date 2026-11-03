import { existsSync, writeFileSync } from 'node:fs';
import { pathToFileURL } from 'node:url';
import type Oway from '@oway/sdk';
import type { QuoteRequest, ShipmentRequest } from '@oway/sdk';
import { inputFile, required, sandboxClient, QuickstartError, reportFailure } from './quickstart-config';

export async function shipperQuickstart(
  oway: Oway, quoteRequest: QuoteRequest, shipmentRequest: ShipmentRequest, idempotencyKey: string,
  save: (state: { request: ShipmentRequest; idempotencyKey: string }) => void = () => {},
  progress: (stage: string) => void = () => {},
): Promise<string> {
  if (!idempotencyKey) throw new QuickstartError('A stable idempotency key is required');
  progress('Request quote');
  const quote = await oway.quotes.create(quoteRequest);
  if (!quote.id) throw new QuickstartError('No usable quote returned; inspect quote eligibility before creating');
  const request = { ...shipmentRequest, quoteId: quote.id };
  progress('Save exact creation for recovery');
  save({ request, idempotencyKey });
  progress('Create shipment');
  const shipment = await oway.shipments.create(request, undefined, { idempotencyKey });
  if (!shipment.orderNumber) throw new QuickstartError('Create returned no orderNumber; reconcile before retrying');
  progress('Replay exact creation');
  const replay = await oway.shipments.create(request, undefined, { idempotencyKey });
  if (replay.orderNumber !== shipment.orderNumber) throw new QuickstartError('Create replay returned a different shipment');
  progress('Confirm shipment');
  const confirmed = await oway.shipments.confirm(shipment.orderNumber);
  if (confirmed.orderStatus !== 'CONFIRMED') throw new QuickstartError('Shipment is not CONFIRMED; inspect confirmation prerequisites');
  progress('Retrieve tracking');
  await oway.shipments.tracking(shipment.orderNumber, { includeLocation: false });
  progress('Retrieve bill of lading');
  const document = await oway.shipments.document(shipment.orderNumber, 'BILL_OF_LADING');
  if (!document.url) throw new QuickstartError('BOL is not available; follow document readiness guidance');
  return shipment.orderNumber;
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  const stateFile = required('OWAY_STATE_FILE');
  if (existsSync(stateFile)) throw new QuickstartError('State file exists; reconcile its saved creation before another run');
  shipperQuickstart(sandboxClient(), inputFile('OWAY_QUOTE_FILE'), inputFile('OWAY_SHIPMENT_FILE'), required('OWAY_IDEMPOTENCY_KEY'),
    state => writeFileSync(stateFile, JSON.stringify(state), { mode: 0o600, flag: 'wx' }),
    stage => console.log(stage))
    .then(orderNumber => console.log(`Sandbox shipment ${orderNumber}: confirmed, tracking and BOL retrieved`))
    .catch((error: unknown) => { reportFailure(error); console.error('Sandbox journey stopped. Reconcile the last completed operation before rerunning; see the quickstart recovery guide.'); process.exitCode = 1; });
}
