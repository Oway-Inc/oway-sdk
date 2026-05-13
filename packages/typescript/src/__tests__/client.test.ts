import { describe, it, expect, vi, beforeEach } from 'vitest';
import { HttpClient, OwayError, parseHttpError } from '../client';

global.fetch = vi.fn();

const okToken = () => ({
  ok: true,
  status: 200,
  headers: new Map([['content-type', 'application/json']]),
  text: async () => JSON.stringify({ accessToken: 'tok', expiresIn: 3600 }),
  json: async () => ({ accessToken: 'tok', expiresIn: 3600 }),
});

const okJson = (body: unknown = {}) => ({
  ok: true,
  status: 200,
  headers: new Map(),
  text: async () => JSON.stringify(body),
  json: async () => body,
});

const errJson = (status: number, body: unknown, requestId = 'req-srv') => ({
  ok: false,
  status,
  headers: new Map([['x-request-id', requestId]]),
  text: async () => (typeof body === 'string' ? body : JSON.stringify(body)),
  json: async () => body,
});

describe('OwayError', () => {
  it('classifies retryable status codes', () => {
    for (const s of [408, 429, 500, 502, 503, 504]) {
      expect(new OwayError({ message: '', statusCode: s }).isRetryable()).toBe(true);
    }
  });

  it('classifies non-retryable status codes', () => {
    for (const s of [400, 401, 403, 404, 422, 501]) {
      expect(new OwayError({ message: '', statusCode: s }).isRetryable()).toBe(false);
    }
  });

  it('exposes client/server bucket helpers', () => {
    expect(new OwayError({ message: '', statusCode: 422 }).isClientError()).toBe(true);
    expect(new OwayError({ message: '', statusCode: 503 }).isServerError()).toBe(true);
  });
});

describe('parseHttpError', () => {
  it('populates all fields from a ProblemDetail body', () => {
    const e = parseHttpError(
      422,
      'req-123',
      JSON.stringify({
        status: 422,
        title: 'Unprocessable Entity',
        detail: 'lane not covered',
        reason: 'no_coverage',
      })
    );
    expect(e.statusCode).toBe(422);
    expect(e.message).toBe('lane not covered');
    expect(e.code).toBe('no_coverage');
    expect(e.requestId).toBe('req-123');
  });

  it('falls back to title when detail is absent', () => {
    const e = parseHttpError(400, undefined, JSON.stringify({ title: 'Bad Request' }));
    expect(e.message).toBe('Bad Request');
  });

  it('parses violations array', () => {
    const e = parseHttpError(
      400,
      undefined,
      JSON.stringify({
        reason: 'validation_failed',
        violations: [
          { field: 'pickupAddress.zipCode', reason: 'must be 5 digits' },
          { field: 'orderComponents[0].weight', reason: 'required' },
        ],
      })
    );
    expect(e.violations).toHaveLength(2);
    expect(e.violations[0].field).toBe('pickupAddress.zipCode');
    expect(e.violations[1].reason).toBe('required');
  });

  it('is tolerant of malformed bodies', () => {
    const e = parseHttpError(500, 'req', '{not json');
    expect(e.statusCode).toBe(500);
    expect(e.rawBody).toBe('{not json');
  });
});

describe('HttpClient', () => {
  let client: HttpClient;

  beforeEach(() => {
    vi.clearAllMocks();
    client = new HttpClient({
      clientId: 'client_test',
      clientSecret: 'secret_test',
      apiKey: 'oway_sk_default',
      maxRetries: 0,
    });
  });

  it('fetches an M2M token on first request', async () => {
    (fetch as any).mockResolvedValueOnce(okToken()).mockResolvedValueOnce(okJson());
    await client.get('/test');
    const tokenCall = (fetch as any).mock.calls[0];
    expect(JSON.parse(tokenCall[1].body)).toEqual({
      clientId: 'client_test',
      clientSecret: 'secret_test',
    });
  });

  it('caches the token across calls', async () => {
    (fetch as any)
      .mockResolvedValueOnce(okToken())
      .mockResolvedValue(okJson());
    await client.get('/a');
    await client.get('/b');
    await client.get('/c');
    const tokenCalls = (fetch as any).mock.calls.filter((c: any[]) =>
      c[0].includes('/v1/auth/token')
    );
    expect(tokenCalls).toHaveLength(1);
  });

  it('rejects when the token response is malformed', async () => {
    (fetch as any).mockResolvedValueOnce({
      ok: true,
      status: 200,
      headers: new Map(),
      text: async () => '{not json',
    });
    await expect(client.get('/a')).rejects.toThrow(OwayError);
  });

  it('rejects when the token response lacks accessToken', async () => {
    (fetch as any).mockResolvedValueOnce({
      ok: true,
      status: 200,
      headers: new Map(),
      text: async () => JSON.stringify({ accessToken: '', expiresIn: 0 }),
    });
    await expect(client.get('/a')).rejects.toThrow(OwayError);
  });

  it('parses a ProblemDetail body into OwayError fields', async () => {
    (fetch as any)
      .mockResolvedValueOnce(okToken())
      .mockResolvedValueOnce(
        errJson(422, { reason: 'no_coverage', detail: 'lane not covered' })
      );
    await expect(client.get('/test')).rejects.toMatchObject({
      statusCode: 422,
      code: 'no_coverage',
      message: 'lane not covered',
      requestId: 'req-srv',
    });
  });

  it('does not retry a non-retryable status', async () => {
    (fetch as any)
      .mockResolvedValueOnce(okToken())
      .mockResolvedValueOnce(errJson(422, { reason: 'no_coverage' }));
    const c = new HttpClient({
      clientId: 'a',
      clientSecret: 'b',
      apiKey: 'k',
      maxRetries: 3,
    });
    await expect(c.get('/test')).rejects.toMatchObject({ statusCode: 422 });
    const apiCalls = (fetch as any).mock.calls.filter(
      (call: any[]) => !call[0].includes('/v1/auth/token')
    );
    expect(apiCalls).toHaveLength(1);
  });

  it('retries transient failures and succeeds', async () => {
    const c = new HttpClient({
      clientId: 'a',
      clientSecret: 'b',
      apiKey: 'k',
      maxRetries: 3,
    });
    (fetch as any)
      .mockResolvedValueOnce(okToken())
      .mockResolvedValueOnce(errJson(503, { reason: 'unavailable' }))
      .mockResolvedValueOnce(errJson(503, { reason: 'unavailable' }))
      .mockResolvedValueOnce(okJson({ ok: true }));
    const result = await c.get<{ ok: boolean }>('/test');
    expect(result).toEqual({ ok: true });
  });

  it('uses companyApiKey when provided', async () => {
    (fetch as any)
      .mockResolvedValueOnce(okToken())
      .mockResolvedValueOnce(okJson());
    await client.get('/test', undefined, 'oway_sk_tenant');
    const apiCall = (fetch as any).mock.calls[1];
    expect(apiCall[1].headers['x-oway-api-key']).toBe('oway_sk_tenant');
  });

  it('generates a unique request id per call', async () => {
    (fetch as any).mockResolvedValueOnce(okToken()).mockResolvedValue(okJson());
    await client.get('/a');
    await client.get('/b');
    const ids = (fetch as any).mock.calls
      .filter((c: any[]) => !c[0].includes('/v1/auth/token'))
      .map((c: any[]) => c[1].headers['x-request-id']);
    expect(new Set(ids).size).toBe(ids.length);
  });
});
