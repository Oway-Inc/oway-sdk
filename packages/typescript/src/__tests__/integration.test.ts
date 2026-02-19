import { describe, it, expect, beforeAll } from 'vitest';
import Oway from '../index';

/**
 * Integration tests against the real Oway API
 * 
 * These tests require:
 * - OWAY_TEST_API_KEY environment variable
 * - Access to sandbox/test API
 * 
 * Run with: OWAY_TEST_API_KEY=oway_sk_test_... npm test
 */

describe('Integration Tests', () => {
  let oway: Oway;

  beforeAll(() => {
    const apiKey = process.env.OWAY_TEST_API_KEY;
    
    if (!apiKey) {
      console.warn('⚠️  Skipping integration tests - OWAY_TEST_API_KEY not set');
      return;
    }

    oway = new Oway({
      apiKey,
      baseUrl: process.env.OWAY_BASE_URL || 'https://sandbox.api.oway.io',
      debug: true,
    });
  });

  it('should request a quote', async () => {
    if (!oway) return; // Skip if no API key

    const quote = await oway.quotes.create({
      origin: {
        addressLine1: '123 Test St',
        city: 'Los Angeles',
        state: 'CA',
        zipCode: '90210',
        country: 'US',
      },
      destination: {
        addressLine1: '456 Test Ave',
        city: 'New York',
        state: 'NY',
        zipCode: '10001',
        country: 'US',
      },
      items: [{
        description: 'Test shipment',
        weight: 100,
        weightUnit: 'LBS',
        quantity: 1,
      }],
    });

    expect(quote).toBeDefined();
    expect(quote.quoteId).toBeDefined();
    expect(quote.totalCharge).toBeGreaterThan(0);
  });

  it('should handle invalid requests gracefully', async () => {
    if (!oway) return;

    await expect(
      oway.quotes.create({
        // Missing required fields
      } as any)
    ).rejects.toThrow();
  });

  it('should retrieve a quote by ID', async () => {
    if (!oway) return;

    // First create a quote
    const created = await oway.quotes.create({
      origin: { zipCode: '90210', country: 'US' },
      destination: { zipCode: '10001', country: 'US' },
      items: [{ weight: 100, weightUnit: 'LBS', quantity: 1 }],
    });

    // Then retrieve it
    const retrieved = await oway.quotes.retrieve(created.quoteId);
    expect(retrieved.quoteId).toBe(created.quoteId);
  });
});
