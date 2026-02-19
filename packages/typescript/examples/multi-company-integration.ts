/**
 * Multi-Company Integration Example
 * For EDI providers, 3PLs, TMS platforms serving multiple shippers
 */

import Oway from '@oway/sdk';

// M2M credentials from Sales Engineering (same for all companies)
const oway = new Oway({
  clientId: process.env.OWAY_M2M_CLIENT_ID!,
  clientSecret: process.env.OWAY_M2M_CLIENT_SECRET!,
  // No default apiKey - provided per-request
  // Defaults to sandbox: https://rest-api.sandbox.oway.io
});

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
    {
      origin: { zipCode: '90210', country: 'US' },
      destination: { zipCode: '10001', country: 'US' },
      items: [{ weight: 500, weightUnit: 'LBS' }],
    },
    customerAPIKeys['customer_001'] // Customer-specific API key
  );
  console.log(`   Quote: ${quote1.id}\n`);

  // Process shipment for Customer 002
  console.log('📦 Processing for Customer 002...');
  const quote2 = await oway.quotes.create(
    {
      origin: { zipCode: '60601', country: 'US' },
      destination: { zipCode: '94102', country: 'US' },
      items: [{ weight: 300, weightUnit: 'LBS' }],
    },
    customerAPIKeys['customer_002'] // Different customer key
  );
  console.log(`   Quote: ${quote2.id}\n`);

  // Process shipment for Customer 003
  console.log('📦 Processing for Customer 003...');
  const quote3 = await oway.quotes.create(
    {
      origin: { zipCode: '75201', country: 'US' },
      destination: { zipCode: '02108', country: 'US' },
      items: [{ weight: 200, weightUnit: 'LBS' }],
    },
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
