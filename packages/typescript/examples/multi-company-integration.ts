/**
 * Multi-Company Integration Example
 * For EDI providers, 3PLs, TMS platforms serving multiple shippers
 */

import Oway, { type QuoteRequest } from '@oway/sdk';

// M2M credentials from Sales Engineering (same for all companies)
const oway = new Oway({
  clientId: process.env.OWAY_M2M_CLIENT_ID!,
  clientSecret: process.env.OWAY_M2M_CLIENT_SECRET!,
  // No default apiKey - provided per-request
  // Defaults to sandbox: https://api.sandbox.oway.io
});

type Stop = { zipCode: string; city: string; state: string };

// A synthetic lane; a real integration builds this from the customer's tender.
function lane(pickup: Stop, delivery: Stop, poundsWeight: number): QuoteRequest {
  const address = (stop: Stop) => ({
    ...stop,
    name: 'Example Facility',
    address1: '100 Example Street',
    phoneNumber: '+15555550100',
    contactPerson: 'Example Contact',
  });
  return {
    pickupAddress: address(pickup),
    deliveryAddress: address(delivery),
    orderComponents: [{ palletCount: 1, poundsWeight }],
  };
}

// Customer API keys from your database
const customerAPIKeys: Record<string, string> = {
  'customer_001': process.env.CUSTOMER_001_API_KEY || 'oway_sk_test_customer001',
  'customer_002': process.env.CUSTOMER_002_API_KEY || 'oway_sk_test_customer002',
  'customer_003': process.env.CUSTOMER_003_API_KEY || 'oway_sk_test_customer003',
};

async function main() {
  console.log('🚚 Multi-Company Integration Example\n');

  // Process shipment for Customer 001
  console.log('📦 Processing for Customer 001...');
  const quote1 = await oway.quotes.create(
    lane(
      { zipCode: '90001', city: 'Los Angeles', state: 'CA' },
      { zipCode: '10001', city: 'New York', state: 'NY' },
      500,
    ),
    customerAPIKeys['customer_001'] // Customer-specific API key
  );
  console.log(`   Quote: ${quote1.id}\n`);

  // Process shipment for Customer 002
  console.log('📦 Processing for Customer 002...');
  const quote2 = await oway.quotes.create(
    lane(
      { zipCode: '60601', city: 'Chicago', state: 'IL' },
      { zipCode: '94102', city: 'San Francisco', state: 'CA' },
      300,
    ),
    customerAPIKeys['customer_002'] // Different customer key
  );
  console.log(`   Quote: ${quote2.id}\n`);

  // Process shipment for Customer 003
  console.log('📦 Processing for Customer 003...');
  const quote3 = await oway.quotes.create(
    lane(
      { zipCode: '75201', city: 'Dallas', state: 'TX' },
      { zipCode: '02108', city: 'Boston', state: 'MA' },
      200,
    ),
    customerAPIKeys['customer_003']
  );
  console.log(`   Quote: ${quote3.id}\n`);

  console.log('✅ All quotes created!');
  console.log('💡 M2M token reused across all customers');
  console.log('🛡️ Running on sandbox - safe for testing');
}

main().catch(error => {
  console.error('❌ Error:', error.message);
  process.exit(1);
});
