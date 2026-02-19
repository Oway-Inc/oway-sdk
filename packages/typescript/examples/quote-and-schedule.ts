/**
 * Complete example: Get a quote and schedule a shipment
 * Demonstrates the full workflow using M2M authentication
 */

import Oway from '@oway/sdk';

async function main() {
  // Initialize with M2M credentials (defaults to sandbox)
  const oway = new Oway({
    clientId: process.env.OWAY_M2M_CLIENT_ID || 'client_example',
    clientSecret: process.env.OWAY_M2M_CLIENT_SECRET || 'secret_example',
    apiKey: process.env.OWAY_API_KEY || 'oway_sk_test_example',
    // Uses sandbox by default: https://rest-api.sandbox.oway.io
  });

  console.log('🚚 Oway SDK Example: Quote and Schedule\n');

  // Step 1: Get a quote
  console.log('📋 Step 1: Requesting quote...');
  const quote = await oway.quotes.create({
    origin: {
      addressLine1: '123 Warehouse Rd',
      city: 'Los Angeles',
      state: 'CA',
      zipCode: '90210',
      country: 'US',
    },
    destination: {
      addressLine1: '456 Distribution Center',
      city: 'New York',
      state: 'NY',
      zipCode: '10001',
      country: 'US',
    },
    items: [{
      description: 'Pallets of goods',
      weight: 500,
      weightUnit: 'LBS',
      quantity: 2,
    }],
  });

  console.log(`✅ Quote received!`);
  console.log(`   Quote ID: ${quote.id}`);
  console.log(`   Price: $${(quote.quotedPriceInCents || 0) / 100}\n`);

  // Step 2: Schedule the shipment
  console.log('📦 Step 2: Scheduling shipment...');
  const shipment = await oway.shipments.create({
    quoteId: quote.id!,
    pickup: {
      contactName: 'John Doe',
      contactPhone: '555-0123',
      contactEmail: 'john@example.com',
    },
    delivery: {
      contactName: 'Jane Smith',
      contactPhone: '555-0456',
      contactEmail: 'jane@example.com',
    },
  });

  console.log(`✅ Shipment created: ${shipment.orderNumber}\n`);

  // Step 3: Confirm the shipment
  console.log('✔️  Step 3: Confirming shipment...');
  await oway.shipments.confirm(shipment.orderNumber!);
  console.log(`✅ Confirmed!\n`);

  // Step 4: Track the shipment
  console.log('📍 Step 4: Tracking...');
  const tracking = await oway.shipments.tracking(shipment.orderNumber!);
  console.log(`   Status: ${tracking.orderStatus || 'Unknown'}\n`);

  console.log('✨ Complete! Running on sandbox - no real shipments created.');
}

main().catch(error => {
  console.error('❌ Error:', error.message);
  process.exit(1);
});
