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
    // Uses sandbox by default: https://api.sandbox.oway.io
  });

  console.log('🚚 Oway SDK Example: Quote and Schedule\n');

  // Step 1: Get a quote
  console.log('📋 Step 1: Requesting quote...');
  const pickupAddress = {
    name: 'Example Warehouse',
    address1: '123 Warehouse Rd',
    city: 'Los Angeles',
    state: 'CA',
    zipCode: '90001',
    phoneNumber: '+12135550123',
    contactPerson: 'Example Shipper',
  };
  const deliveryAddress = {
    name: 'Example Distribution Center',
    address1: '456 Distribution Ave',
    city: 'New York',
    state: 'NY',
    zipCode: '10001',
    phoneNumber: '+12125550456',
    contactPerson: 'Example Receiver',
  };
  const orderComponents = [
    { palletCount: 2, poundsWeight: 1000, dimensions: { length: 48, width: 40, height: 48 } },
  ];
  const quote = await oway.quotes.create({ pickupAddress, deliveryAddress, orderComponents });

  console.log(`✅ Quote received!`);
  console.log(`   Quote ID: ${quote.id}`);
  console.log(`   Price: $${(quote.quotedPriceInCents || 0) / 100}\n`);

  // Step 2: Schedule the shipment
  console.log('📦 Step 2: Scheduling shipment...');
  const shipment = await oway.shipments.create({
    quoteId: quote.id,
    pickupAddress,
    deliveryAddress,
    orderComponents,
    description: 'Palletized freight',
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
