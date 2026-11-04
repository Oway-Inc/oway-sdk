import { existsSync, readFileSync, writeFileSync } from 'node:fs';
import { basename } from 'node:path';
import { pathToFileURL } from 'node:url';
import type Oway from '@oway/sdk';
import type { UploadConfirmationRequest, OfferLocationUpdate } from '@oway/sdk';
import { inputFile, required, sandboxClient, QuickstartError, reportFailure } from './quickstart-config';

export interface CarrierRun {
  offerId: string;
  documentId: string;
  metadata: UploadConfirmationRequest;
}

export async function prepareCarrier(
  oway: Oway, offerId: string, file: Uint8Array, metadata: UploadConfirmationRequest,
  location: OfferLocationUpdate, save: (run: CarrierRun) => void,
  upload: (url: string, contentType: string, bytes: Uint8Array) => Promise<void>,
  progress: (stage: string) => void = () => {},
): Promise<CarrierRun> {
  if (!file.length || file.length > 10 * 1024 * 1024) throw new QuickstartError('POD must contain 1 byte to 10 MiB');
  progress('Retrieve assigned offer');
  let offer = await oway.carrier.offer(offerId);
  if (offer.status === 'PENDING') { progress('Accept offer'); offer = await oway.carrier.acceptOffer(offerId, {}); }
  if (offer.status !== 'ACCEPTED' && offer.status !== 'PICKED_UP') throw new QuickstartError('Use an assigned pending or accepted sandbox offer');
  if (offer.status === 'ACCEPTED') {
    progress('Confirm pickup');
    const pickedUp = await oway.carrier.confirmPickup(offerId, { pickup_time: new Date().toISOString() });
    if (pickedUp.status !== 'PICKED_UP') throw new QuickstartError('Pickup did not reach PICKED_UP');
  }
  progress('Submit location');
  await oway.carrier.submitLocation(offerId, location);
  progress('Request upload URL');
  const staging = await oway.carrier.requestUploadUrl(offerId, metadata);
  if (!staging.document_id || !staging.upload_url || staging.content_type !== metadata.content_type) throw new QuickstartError('Incomplete upload response');
  if (new URL(staging.upload_url).protocol !== 'https:') throw new QuickstartError('Expected HTTPS upload URL');
  const run = { offerId, documentId: staging.document_id, metadata };
  progress('Save upload identity for recovery');
  save(run);
  progress('Upload POD bytes');
  await upload(staging.upload_url, metadata.content_type, file);
  progress('Complete upload registration');
  const completed = await oway.carrier.confirmUpload(offerId, run.documentId, metadata);
  if (completed.document_id !== run.documentId) throw new QuickstartError('Completion returned a different document');
  return run;
}

export async function deliverCarrier(oway: Oway, run: CarrierRun, progress: (stage: string) => void = () => {}): Promise<void> {
  progress('Check POD review');
  const document = await oway.carrier.uploadedDocument(run.offerId, run.documentId);
  if (document.review_status !== 'APPROVED') throw new QuickstartError('POD review is pending; resume after authorized approval with a verified recipient signature');
  progress('Reconcile shipment status');
  const current = await oway.carrier.shipment(run.offerId);
  if (current.status === 'DELIVERED') return;
  if (current.status !== 'PICKED_UP') throw new QuickstartError('Shipment must be picked up before delivery');
  progress('Confirm delivery');
  const delivered = await oway.carrier.confirmDelivery(run.offerId, {
    delivery_time: new Date().toISOString(),
    proof_of_delivery_document_id: run.documentId,
  });
  if (delivered.status !== 'DELIVERED') throw new QuickstartError('Shipment did not reach DELIVERED');
  progress('Retrieve registered POD');
  await oway.carrier.uploadedDocument(run.offerId, run.documentId);
}

async function main() {
  const client = sandboxClient();
  const stateFile = required('OWAY_STATE_FILE');
  const step = required('OWAY_STEP');
  if (step === 'prepare') {
    if (existsSync(stateFile)) throw new QuickstartError('State file already exists; reconcile it before preparing another upload');
    const filename = required('OWAY_POD_FILE');
    await prepareCarrier(client, required('OWAY_OFFER_ID'), readFileSync(filename), {
      document_type: 'PROOF_OF_DELIVERY', content_type: 'application/pdf', file_name: basename(filename),
    }, { ...inputFile<OfferLocationUpdate>('OWAY_LOCATION_FILE'), timestamp: new Date().toISOString() },
    run => writeFileSync(stateFile, JSON.stringify(run), { mode: 0o600, flag: 'wx' }),
    async (url, contentType, bytes) => {
      const result = await fetch(url, { method: 'PUT', headers: { 'Content-Type': contentType }, body: Buffer.from(bytes), redirect: 'error' });
      if (!result.ok) throw new QuickstartError('Storage PUT failed; preserve the state and follow the new-upload recovery steps');
    }, stage => console.log(stage));
    console.log('POD registered. Request authorized review, then run OWAY_STEP=deliver. State saved without signed URLs.');
  } else {
    const run = JSON.parse(readFileSync(stateFile, 'utf8')) as CarrierRun;
    if (step === 'confirm-upload') {
      console.log('Reconcile saved upload completion');
      await client.carrier.confirmUpload(run.offerId, run.documentId, run.metadata);
      console.log('Upload completion reconciled; review may still be pending.');
    } else if (step === 'deliver') {
      await deliverCarrier(client, run, stage => console.log(stage));
      console.log('Sandbox shipment delivered; registered POD retrieved.');
    } else throw new QuickstartError('OWAY_STEP must be prepare, confirm-upload or deliver');
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  main().catch((error: unknown) => { reportFailure(error); console.error('Carrier journey stopped. Preserve the state file and reconcile the last operation; see recovery instructions.'); process.exitCode = 1; });
}
