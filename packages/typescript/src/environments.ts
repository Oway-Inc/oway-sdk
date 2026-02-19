/**
 * Oway API environment URLs
 * Use these constants instead of hardcoding URLs
 */

export const OwayEnvironments = {
  /**
   * Sandbox environment for development and testing
   * Safe to use - no real shipments created
   */
  SANDBOX: 'https://rest-api.sandbox.oway.io',

  /**
   * Production environment for live traffic
   * Real shipments will be created and billed
   */
  PRODUCTION: 'https://rest-api.oway.io',
} as const;

export type OwayEnvironment = typeof OwayEnvironments[keyof typeof OwayEnvironments];
