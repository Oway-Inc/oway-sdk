import type { components } from './generated/schema';

/** Token response from the /v1/auth/token endpoint (generated from OpenAPI spec) */
export type TokenResponse = components['schemas']['TokenResponse'];

/** Token error response from the /v1/auth/token endpoint (generated from OpenAPI spec) */
export type TokenErrorResponse = components['schemas']['TokenErrorResponse'];

export interface OwayConfig {
  /**
   * M2M Client ID (REQUIRED for all integrations)
   * Provided by Oway Sales Engineering team
   */
  clientId: string;

  /**
   * M2M Client Secret (REQUIRED for all integrations)
   * Provided by Oway Sales Engineering team
   */
  clientSecret: string;

  /**
   * Default company API key (optional)
   *
   * Single-company: Set this to your company's API key
   * Multi-company: Omit this and provide per-request
   *
   * Get from: https://app.oway.io/settings/api
   */
  apiKey?: string;

  /**
   * Base URL for the Oway API
   * @default "https://rest-api.sandbox.oway.io"
   */
  baseUrl?: string;

  /**
   * Token endpoint for authentication
   * @default "https://rest-api.sandbox.oway.io/v1/auth/token"
   */
  tokenUrl?: string;

  /**
   * Maximum number of retry attempts for failed requests
   * @default 3
   */
  maxRetries?: number;

  /**
   * Timeout in milliseconds for API requests
   * @default 30000
   */
  timeout?: number;

  /**
   * Enable debug logging (logs sanitized request/response metadata)
   * @default false
   */
  debug?: boolean;

  /**
   * Custom logger implementation
   */
  logger?: {
    debug: (msg: string, meta?: Record<string, any>) => void;
    info: (msg: string, meta?: Record<string, any>) => void;
    warn: (msg: string, meta?: Record<string, any>) => void;
    error: (msg: string, meta?: Record<string, any>) => void;
  };
}

export class OwayError extends Error {
  constructor(
    message: string,
    public code?: string,
    public statusCode?: number,
    public requestId?: string
  ) {
    super(message);
    this.name = 'OwayError';
  }

  /**
   * Determines if this error represents a transient failure that should be retried
   */
  isRetryable(): boolean {
    if (!this.statusCode) return false;

    // 429 Too Many Requests - rate limit (retryable with backoff)
    if (this.statusCode === 429) return true;

    // 503 Service Unavailable - temporary outage
    if (this.statusCode === 503) return true;

    // 500 Internal Server Error - might be transient
    if (this.statusCode === 500) return true;

    // 501 Not Implemented - permanent error
    if (this.statusCode === 501) return false;

    // 502 Bad Gateway - might be transient
    if (this.statusCode === 502) return true;

    // 504 Gateway Timeout - might be transient
    if (this.statusCode === 504) return true;

    // Other 5xx errors - retry by default
    if (this.statusCode >= 500) return true;

    // 4xx errors (except 429) are client errors - don't retry
    return false;
  }
}

/**
 * HTTP client for making authenticated requests to the Oway API
 */
export class HttpClient {
  private config: Required<Omit<OwayConfig, 'logger' | 'apiKey' | 'clientId' | 'clientSecret' | 'companyApiKey'>> & {
    apiKey?: string;
    clientId?: string;
    clientSecret?: string;
    companyApiKey?: string;
    logger?: OwayConfig['logger'];
  };
  private accessToken: string | null = null;
  private tokenExpiry: number = 0;
  private tokenRefreshPromise: Promise<string> | null = null;

  constructor(config: OwayConfig) {
    // M2M credentials are REQUIRED for all integrations
    if (!config.clientId || !config.clientSecret) {
      throw new OwayError('clientId and clientSecret are required. Contact Oway Sales Engineering to obtain M2M credentials.');
    }

    this.config = {
      baseUrl: config.baseUrl || process.env.OWAY_BASE_URL || 'https://rest-api.sandbox.oway.io',
      tokenUrl: config.tokenUrl || (config.baseUrl || process.env.OWAY_BASE_URL || 'https://rest-api.sandbox.oway.io') + '/v1/auth/token',
      maxRetries: config.maxRetries ?? 3,
      timeout: config.timeout ?? 30000,
      debug: config.debug ?? false,
      clientId: config.clientId,
      clientSecret: config.clientSecret,
      apiKey: config.apiKey, // Optional default company API key
      logger: config.logger,
    };

    this.log('debug', 'Oway SDK initialized', {
      baseUrl: this.config.baseUrl,
      authMode: 'M2M',
      hasDefaultApiKey: !!this.config.apiKey,
    });
  }

  /**
   * Internal logging with sanitization
   */
  private log(
    level: 'debug' | 'info' | 'warn' | 'error',
    message: string,
    meta?: Record<string, any>
  ): void {
    if (!this.config.debug && level === 'debug') return;

    const sanitized = meta ? this.sanitizeForLogging(meta) : undefined;

    if (this.config.logger) {
      this.config.logger[level](message, sanitized);
    } else if (this.config.debug && level !== 'debug') {
      // Default console logging for non-debug levels when debug enabled
      console[level === 'error' ? 'error' : 'log'](`[Oway ${level}]`, message, sanitized);
    }
  }

  /**
   * Sanitize objects for logging - remove sensitive fields
   */
  private sanitizeForLogging(obj: any): any {
    if (!obj || typeof obj !== 'object') return obj;

    const sensitive = ['apiKey', 'token', 'authorization', 'password', 'secret'];
    const sanitized: any = Array.isArray(obj) ? [] : {};

    for (const [key, value] of Object.entries(obj)) {
      const lowerKey = key.toLowerCase();
      if (sensitive.some(s => lowerKey.includes(s))) {
        sanitized[key] = '[REDACTED]';
      } else if (value && typeof value === 'object') {
        sanitized[key] = this.sanitizeForLogging(value);
      } else {
        sanitized[key] = value;
      }
    }

    return sanitized;
  }

  /**
   * Get or refresh the access token using the API key
   * Handles concurrent requests by queuing them behind a single refresh
   */
  private async getAccessToken(): Promise<string> {
    // If refresh already in progress, wait for it
    if (this.tokenRefreshPromise) {
      this.log('debug', 'Waiting for token refresh in progress');
      return this.tokenRefreshPromise;
    }

    // Return cached token if still valid (with 5-minute buffer)
    if (this.accessToken && Date.now() < this.tokenExpiry - 5 * 60 * 1000) {
      return this.accessToken;
    }

    // Start token refresh
    this.log('debug', 'Refreshing access token');
    this.tokenRefreshPromise = this.refreshToken();

    try {
      this.accessToken = await this.tokenRefreshPromise;
      this.log('info', 'Access token refreshed', {
        expiresAt: new Date(this.tokenExpiry).toISOString(),
      });
      return this.accessToken;
    } finally {
      this.tokenRefreshPromise = null;
    }
  }

  /**
   * Perform the actual token refresh using M2M credentials
   */
  private async refreshToken(): Promise<string> {
    try {
      this.log('debug', 'Refreshing M2M access token');

      const response = await fetch(this.config.tokenUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          clientId: this.config.clientId,
          clientSecret: this.config.clientSecret,
        }),
      });

      if (!response.ok) {
        throw new OwayError(
          'Failed to obtain access token',
          'AUTH_FAILED',
          response.status
        );
      }

      const data = await response.json() as TokenResponse;

      if (!data.accessToken || !data.expiresIn) {
        throw new OwayError(
          'Invalid token response: missing accessToken or expiresIn',
          'AUTH_INVALID_RESPONSE'
        );
      }

      this.tokenExpiry = Date.now() + data.expiresIn * 1000;

      return data.accessToken;
    } catch (error) {
      this.log('error', 'Token refresh failed', {
        error: error instanceof Error ? error.message : 'Unknown error',
      });

      if (error instanceof OwayError) throw error;
      throw new OwayError(
        `Authentication failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'AUTH_ERROR'
      );
    }
  }

  /**
   * Generate a unique request ID
   */
  private generateRequestId(): string {
    // Simple UUID v4-like implementation
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
      const r = (Math.random() * 16) | 0;
      const v = c === 'x' ? r : (r & 0x3) | 0x8;
      return v.toString(16);
    });
  }

  /**
   * Make an authenticated request to the Oway API
   */
  async request<T>(
    method: string,
    path: string,
    options: {
      query?: Record<string, string | number | boolean>;
      body?: unknown;
      headers?: Record<string, string>;
      requestId?: string;
      companyApiKey?: string; // Override per-request for multi-tenant integrations
    } = {}
  ): Promise<T> {
    const token = await this.getAccessToken();
    const url = new URL(path, this.config.baseUrl);
    const requestId = options.requestId || this.generateRequestId();

    // Add query parameters
    if (options.query) {
      Object.entries(options.query).forEach(([key, value]) => {
        url.searchParams.append(key, String(value));
      });
    }

    // Determine which API key to use (priority: per-request > default > fallback to auth apiKey)
    const apiKey = options.companyApiKey || this.config.companyApiKey || this.config.apiKey;

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      'x-request-id': requestId,
      ...options.headers,
    };

    // Add company API key if available
    if (apiKey) {
      headers['x-oway-api-key'] = apiKey;
    }

    this.log('debug', `${method} ${path}`, {
      requestId,
      hasBody: !!options.body,
      query: options.query,
    });

    let lastError: Error | null = null;

    for (let attempt = 0; attempt <= this.config.maxRetries; attempt++) {
      try {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.config.timeout);

        const response = await fetch(url.toString(), {
          method,
          headers,
          body: options.body ? JSON.stringify(options.body) : undefined,
          signal: controller.signal,
        });

        clearTimeout(timeoutId);

        // Extract server request ID (if different from ours)
        const serverRequestId = response.headers.get('x-request-id') || requestId;

        if (!response.ok) {
          let errorData: any;
          try {
            errorData = await response.json();
          } catch {
            errorData = { message: response.statusText };
          }

          const error = new OwayError(
            errorData.message || `Request failed with status ${response.status}`,
            errorData.code || 'API_ERROR',
            response.status,
            serverRequestId
          );

          this.log('warn', 'Request failed', {
            method,
            path,
            status: response.status,
            requestId: serverRequestId,
            attempt: attempt + 1,
            isRetryable: error.isRetryable(),
          });

          throw error;
        }

        this.log('debug', 'Request successful', {
          method,
          path,
          status: response.status,
          requestId: serverRequestId,
        });

        // Return empty object for 204 No Content
        if (response.status === 204) {
          return {} as T;
        }

        return await response.json() as T;
      } catch (error) {
        lastError = error as Error;

        // Use isRetryable() for intelligent retry decisions
        if (error instanceof OwayError && !error.isRetryable()) {
          this.log('error', 'Non-retryable error', {
            requestId,
            code: error.code,
            statusCode: error.statusCode,
          });
          throw error;
        }

        // Don't retry if this was the last attempt
        if (attempt === this.config.maxRetries) {
          this.log('error', 'Max retries exceeded', {
            requestId,
            attempts: attempt + 1,
          });
          break;
        }

        // Exponential backoff: 1s, 2s, 4s
        const delay = Math.pow(2, attempt) * 1000;
        this.log('warn', 'Retrying request', {
          requestId,
          attempt: attempt + 1,
          maxRetries: this.config.maxRetries,
          delayMs: delay,
        });
        await new Promise(resolve => setTimeout(resolve, delay));
      }
    }

    this.log('error', 'Request failed', {
      requestId,
      error: lastError instanceof Error ? lastError.message : 'Unknown error',
    });

    throw lastError || new OwayError('Request failed after retries', 'MAX_RETRIES_EXCEEDED', undefined, requestId);
  }

  async get<T>(path: string, query?: Record<string, string | number | boolean>, companyApiKey?: string): Promise<T> {
    return this.request<T>('GET', path, { query, companyApiKey });
  }

  async post<T>(path: string, body?: unknown, companyApiKey?: string): Promise<T> {
    return this.request<T>('POST', path, { body, companyApiKey });
  }

  async put<T>(path: string, body?: unknown, companyApiKey?: string): Promise<T> {
    return this.request<T>('PUT', path, { body, companyApiKey });
  }

  async delete<T>(path: string, companyApiKey?: string): Promise<T> {
    return this.request<T>('DELETE', path, { companyApiKey });
  }
}
