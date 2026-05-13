import type { components } from './generated/schema';

/** Token response from the /v1/auth/token endpoint (generated from OpenAPI spec) */
export type TokenResponse = components['schemas']['TokenResponse'];

/** Token error response from the /v1/auth/token endpoint (generated from OpenAPI spec) */
export type TokenErrorResponse = components['schemas']['TokenErrorResponse'];

/** RFC 9457 Problem Details payload returned by the Oway API on error. */
export type ProblemDetail = components['schemas']['ProblemDetail'];

/** A single field-level validation failure. */
export type Violation = components['schemas']['Violation'];

export interface OwayConfig {
  /** M2M Client ID (required). */
  clientId: string;
  /** M2M Client Secret (required). */
  clientSecret: string;
  /**
   * Default company API key. For multi-tenant integrations pass a
   * per-request key on the `companyApiKey` option instead.
   */
  apiKey?: string;
  /** Base URL. Defaults to https://api.sandbox.oway.io. */
  baseUrl?: string;
  /** Token endpoint. Defaults to `${baseUrl}/v1/auth/token`. */
  tokenUrl?: string;
  /** Maximum retry attempts on transient errors. Defaults to 3. Pass 0 to disable. */
  maxRetries?: number;
  /** Request timeout in milliseconds. Defaults to 30000. */
  timeout?: number;
  /** Enable structured debug logging. */
  debug?: boolean;
  /** Custom logger. When omitted the SDK is silent. */
  logger?: {
    debug: (msg: string, meta?: Record<string, unknown>) => void;
    info: (msg: string, meta?: Record<string, unknown>) => void;
    warn: (msg: string, meta?: Record<string, unknown>) => void;
    error: (msg: string, meta?: Record<string, unknown>) => void;
  };
}

/**
 * Typed error returned by the SDK whenever the API returns a non-2xx
 * response. Carries the parsed ProblemDetail (`code`, `message`, `violations`)
 * plus the `requestId` to quote when reporting issues.
 */
export class OwayError extends Error {
  readonly statusCode?: number;
  readonly code?: string;
  readonly requestId?: string;
  readonly violations: Violation[];
  readonly rawBody?: string;

  constructor(opts: {
    message: string;
    statusCode?: number;
    code?: string;
    requestId?: string;
    violations?: Violation[];
    rawBody?: string;
  }) {
    super(opts.message);
    this.name = 'OwayError';
    this.statusCode = opts.statusCode;
    this.code = opts.code;
    this.requestId = opts.requestId;
    this.violations = opts.violations ?? [];
    this.rawBody = opts.rawBody;
  }

  /** True for well-known transient HTTP status codes (408/429/500/502/503/504). */
  isRetryable(): boolean {
    if (!this.statusCode) return false;
    return [408, 429, 500, 502, 503, 504].includes(this.statusCode);
  }

  /** 4xx response. */
  isClientError(): boolean {
    return !!this.statusCode && this.statusCode >= 400 && this.statusCode < 500;
  }

  /** 5xx response. */
  isServerError(): boolean {
    return !!this.statusCode && this.statusCode >= 500 && this.statusCode < 600;
  }
}

/** Parse a raw response body into an OwayError. Tolerant of malformed JSON. */
export function parseHttpError(
  status: number,
  requestId: string | undefined,
  rawBody: string
): OwayError {
  let problem: ProblemDetail | undefined;
  try {
    problem = rawBody ? (JSON.parse(rawBody) as ProblemDetail) : undefined;
  } catch {
    // fall through
  }
  const message =
    problem?.detail ||
    problem?.title ||
    `Request failed with status ${status}`;
  return new OwayError({
    message,
    statusCode: status,
    code: problem?.reason,
    requestId,
    violations: problem?.violations,
    rawBody,
  });
}

function randomRequestId(): string {
  if (typeof globalThis !== 'undefined' && globalThis.crypto?.randomUUID) {
    return globalThis.crypto.randomUUID();
  }
  // Fallback for older runtimes: 128 random bits as hex.
  const buf = new Uint8Array(16);
  for (let i = 0; i < 16; i++) buf[i] = Math.floor(Math.random() * 256);
  return Array.from(buf, (b) => b.toString(16).padStart(2, '0')).join('');
}

/** Full-jitter exponential backoff capped at 30s. */
function backoffDelayMs(attempt: number): number {
  const base = Math.min(Math.pow(2, attempt) * 1000, 30_000);
  return Math.floor(Math.random() * (base + 1));
}

interface RequestOptions {
  query?: Record<string, string | number | boolean>;
  body?: unknown;
  headers?: Record<string, string>;
  requestId?: string;
  /** Override the default API key for this request (multi-tenant integrations). */
  companyApiKey?: string;
}

export class HttpClient {
  private readonly config: Required<
    Pick<OwayConfig, 'baseUrl' | 'tokenUrl' | 'maxRetries' | 'timeout' | 'debug' | 'clientId' | 'clientSecret'>
  > & {
    apiKey?: string;
    logger?: OwayConfig['logger'];
  };
  private accessToken: string | null = null;
  private tokenExpiry = 0;
  private tokenRefreshPromise: Promise<string> | null = null;

  constructor(config: OwayConfig) {
    if (!config.clientId || !config.clientSecret) {
      throw new OwayError({
        message: 'clientId and clientSecret are required',
        code: 'CONFIG_MISSING_CREDENTIALS',
      });
    }
    const baseUrl =
      config.baseUrl ||
      (typeof process !== 'undefined' ? process.env?.OWAY_BASE_URL : undefined) ||
      'https://api.sandbox.oway.io';

    this.config = {
      baseUrl,
      tokenUrl: config.tokenUrl || `${baseUrl}/v1/auth/token`,
      maxRetries: config.maxRetries ?? 3,
      timeout: config.timeout ?? 30_000,
      debug: config.debug ?? false,
      clientId: config.clientId,
      clientSecret: config.clientSecret,
      apiKey: config.apiKey,
      logger: config.logger,
    };

    this.log('debug', 'sdk initialized', {
      baseUrl: this.config.baseUrl,
      hasDefaultApiKey: !!this.config.apiKey,
    });
  }

  private log(
    level: 'debug' | 'info' | 'warn' | 'error',
    message: string,
    meta?: Record<string, unknown>
  ): void {
    if (!this.config.logger) return;
    if (!this.config.debug && level === 'debug') return;
    const safe = meta ? (this.sanitize(meta) as Record<string, unknown>) : undefined;
    this.config.logger[level](message, safe);
  }

  private sanitize(obj: unknown): unknown {
    if (!obj || typeof obj !== 'object') return obj;
    const sensitive = ['apikey', 'token', 'authorization', 'password', 'secret', 'clientsecret'];
    const out: Record<string, unknown> = Array.isArray(obj) ? ([] as unknown as Record<string, unknown>) : {};
    for (const [k, v] of Object.entries(obj as Record<string, unknown>)) {
      const lower = k.toLowerCase();
      if (sensitive.some((s) => lower.includes(s))) {
        out[k] = '[REDACTED]';
      } else if (v && typeof v === 'object') {
        out[k] = this.sanitize(v);
      } else {
        out[k] = v;
      }
    }
    return out;
  }

  private async getAccessToken(): Promise<string> {
    if (this.tokenRefreshPromise) return this.tokenRefreshPromise;
    if (this.accessToken && Date.now() < this.tokenExpiry - 5 * 60 * 1000) {
      return this.accessToken;
    }
    this.tokenRefreshPromise = this.refreshToken();
    try {
      this.accessToken = await this.tokenRefreshPromise;
      return this.accessToken;
    } finally {
      this.tokenRefreshPromise = null;
    }
  }

  private async refreshToken(): Promise<string> {
    this.log('debug', 'refreshing access token');
    const resp = await fetch(this.config.tokenUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        clientId: this.config.clientId,
        clientSecret: this.config.clientSecret,
      }),
    });
    const text = await resp.text();
    if (!resp.ok) {
      throw parseHttpError(resp.status, resp.headers.get('x-request-id') ?? undefined, text);
    }
    let parsed: TokenResponse;
    try {
      parsed = JSON.parse(text) as TokenResponse;
    } catch {
      throw new OwayError({
        message: 'token response was not valid JSON',
        code: 'AUTH_INVALID_RESPONSE',
        statusCode: resp.status,
      });
    }
    if (!parsed.accessToken || !parsed.expiresIn) {
      throw new OwayError({
        message: 'token response missing accessToken or expiresIn',
        code: 'AUTH_INVALID_RESPONSE',
        statusCode: resp.status,
      });
    }
    this.tokenExpiry = Date.now() + parsed.expiresIn * 1000;
    return parsed.accessToken;
  }

  async request<T>(method: string, path: string, options: RequestOptions = {}): Promise<T> {
    const token = await this.getAccessToken();
    const url = new URL(path, this.config.baseUrl);
    if (options.query) {
      for (const [k, v] of Object.entries(options.query)) url.searchParams.append(k, String(v));
    }
    const requestId = options.requestId || randomRequestId();
    const apiKey = options.companyApiKey ?? this.config.apiKey;
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
      'x-request-id': requestId,
      ...options.headers,
    };
    if (apiKey) headers['x-oway-api-key'] = apiKey;

    let lastErr: Error | null = null;
    for (let attempt = 0; attempt <= this.config.maxRetries; attempt++) {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this.config.timeout);

      try {
        const resp = await fetch(url.toString(), {
          method,
          headers,
          body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
          signal: controller.signal,
        });
        clearTimeout(timeoutId);

        const serverRequestId = resp.headers.get('x-request-id') ?? requestId;

        if (!resp.ok) {
          const rawBody = await resp.text();
          const err = parseHttpError(resp.status, serverRequestId, rawBody);
          this.log('warn', 'request failed', {
            method,
            path,
            status: err.statusCode,
            code: err.code,
            requestId: err.requestId,
            attempt: attempt + 1,
            retryable: err.isRetryable(),
          });
          if (!err.isRetryable() || attempt === this.config.maxRetries) throw err;
          lastErr = err;
        } else {
          if (resp.status === 204) return {} as T;
          return (await resp.json()) as T;
        }
      } catch (e) {
        clearTimeout(timeoutId);
        if (e instanceof OwayError) {
          if (!e.isRetryable() || attempt === this.config.maxRetries) throw e;
          lastErr = e;
        } else {
          lastErr = e as Error;
          if (attempt === this.config.maxRetries) throw lastErr;
        }
      }

      const delay = backoffDelayMs(attempt);
      this.log('warn', 'retrying', { attempt: attempt + 1, delay });
      await new Promise((r) => setTimeout(r, delay));
    }

    throw (
      lastErr ||
      new OwayError({
        message: 'request failed after retries',
        code: 'MAX_RETRIES_EXCEEDED',
        requestId,
      })
    );
  }

  get<T>(path: string, query?: Record<string, string | number | boolean>, companyApiKey?: string): Promise<T> {
    return this.request<T>('GET', path, { query, companyApiKey });
  }

  post<T>(path: string, body?: unknown, companyApiKey?: string): Promise<T> {
    return this.request<T>('POST', path, { body, companyApiKey });
  }

  put<T>(path: string, body?: unknown, companyApiKey?: string): Promise<T> {
    return this.request<T>('PUT', path, { body, companyApiKey });
  }

  delete<T>(path: string, companyApiKey?: string): Promise<T> {
    return this.request<T>('DELETE', path, { companyApiKey });
  }
}
