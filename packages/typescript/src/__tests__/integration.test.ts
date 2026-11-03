import { readFileSync } from 'node:fs';
import { describe, it, expect, beforeAll } from 'vitest';
import Oway, { OwayEnvironments, type QuoteRequest } from '../index';

const enabled = Boolean(process.env.OWAY_TEST_API_KEY && process.env.OWAY_M2M_CLIENT_ID
  && process.env.OWAY_M2M_CLIENT_SECRET && process.env.OWAY_QUOTE_FILE);

describe.skipIf(!enabled)('Sandbox integration', () => {
  let oway: Oway;
  let request: QuoteRequest;
  beforeAll(() => {
    request = JSON.parse(readFileSync(process.env.OWAY_QUOTE_FILE!, 'utf8'));
    oway = new Oway({
      apiKey: process.env.OWAY_TEST_API_KEY,
      clientId: process.env.OWAY_M2M_CLIENT_ID,
      clientSecret: process.env.OWAY_M2M_CLIENT_SECRET,
      baseUrl: OwayEnvironments.SANDBOX,
      maxRetries: 0,
    });
  });

  it('requests a quote for the provisioned scenario', async () => {
    const quote = await oway.quotes.create(request);
    expect(quote.id).toBeTruthy();
  });

  it('rejects a request missing required fields', async () => {
    await expect(oway.quotes.create({} as QuoteRequest)).rejects.toThrow();
  });

  it('retrieves the quote by returned ID', async () => {
    const created = await oway.quotes.create(request);
    expect(created.id).toBeTruthy();
    const retrieved = await oway.quotes.retrieve(created.id!);
    expect(retrieved.id).toBe(created.id);
  });
});
