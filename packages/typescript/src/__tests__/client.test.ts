import { describe, it, expect, vi, beforeEach } from 'vitest';
import { HttpClient, OwayError } from '../client';

global.fetch = vi.fn();

describe('HttpClient', () => {
  let client: HttpClient;

  beforeEach(() => {
    vi.clearAllMocks();
    client = new HttpClient({
      clientId: 'client_test_123',
      clientSecret: 'secret_test_xyz',
      apiKey: 'oway_sk_test_default',
      debug: false,
    });
  });

  describe('Authentication', () => {
    it('should fetch M2M token on first request', async () => {
      (fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ access_token: 'test_token', expires_in: 3600 }),
      });
      (fetch as any).mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map(),
        json: async () => ({}),
      });

      await client.get('/test');

      const tokenCall = (fetch as any).mock.calls[0];
      expect(JSON.parse(tokenCall[1].body)).toEqual({
        clientId: 'client_test_123',
        clientSecret: 'secret_test_xyz',
      });
    });
  });

  describe('Error Handling', () => {
    it('should classify 500 as retryable', () => {
      const error = new OwayError('Error', 'CODE', 500);
      expect(error.isRetryable()).toBe(true);
    });

    it('should classify 400 as non-retryable', () => {
      const error = new OwayError('Error', 'CODE', 400);
      expect(error.isRetryable()).toBe(false);
    });

    it('should classify 429 as retryable', () => {
      const error = new OwayError('Error', 'CODE', 429);
      expect(error.isRetryable()).toBe(true);
    });
  });
});
