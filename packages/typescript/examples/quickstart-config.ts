import { readFileSync } from 'node:fs';
import Oway, { OwayEnvironments, OwayError } from '@oway/sdk';

export class QuickstartError extends Error {}

export function reportFailure(error: unknown): void {
  if (error instanceof QuickstartError) console.error(error.message);
  else if (error instanceof OwayError) console.error(`API status: ${error.statusCode ?? 'unavailable'}`);
}

export function required(name: string): string {
  const value = process.env[name];
  if (!value) throw new QuickstartError(`Set ${name} before running this sandbox example`);
  return value;
}

export function sandboxClient(): Oway {
  return new Oway({
    clientId: required('OWAY_M2M_CLIENT_ID'),
    clientSecret: required('OWAY_M2M_CLIENT_SECRET'),
    apiKey: required('OWAY_API_KEY'),
    baseUrl: OwayEnvironments.SANDBOX,
    maxRetries: 0,
  });
}

export function inputFile<T>(name: string): T {
  return JSON.parse(readFileSync(required(name), 'utf8')) as T;
}
