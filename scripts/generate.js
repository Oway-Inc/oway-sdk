#!/usr/bin/env node

/**
 * SDK Generation Pipeline
 *
 * Fetches the OpenAPI spec and regenerates all SDK clients.
 *
 * Usage:
 *   node scripts/generate.js                    # fetch from local (localhost:8181)
 *   node scripts/generate.js --sandbox          # fetch from sandbox
 *   node scripts/generate.js --production       # fetch from production
 *   node scripts/generate.js --url <url>        # fetch from custom URL
 *   node scripts/generate.js --skip-fetch       # skip fetch, regenerate from existing spec
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const https = require('https');
const http = require('http');

const ROOT = path.resolve(__dirname, '..');
const SPEC_PATH = path.join(ROOT, 'openapi', 'spec.json');

const URLS = {
  local: 'http://localhost:8181/api-docs/all-v1',
  sandbox: 'https://api.sandbox.oway.io/api-docs/all-v1',
  production: 'https://api.oway.io/api-docs/all-v1',
};

const OAPI_CODEGEN_VERSION = 'v2.5.1';

function parseArgs() {
  const args = process.argv.slice(2);
  if (args.includes('--skip-fetch')) return { skipFetch: true };
  if (args.includes('--production')) return { url: URLS.production };
  if (args.includes('--sandbox')) return { url: URLS.sandbox };
  const urlIdx = args.indexOf('--url');
  if (urlIdx !== -1 && args[urlIdx + 1]) return { url: args[urlIdx + 1] };
  return { url: URLS.local };
}

function fetch(url) {
  return new Promise((resolve, reject) => {
    const client = url.startsWith('https') ? https : http;
    client.get(url, (res) => {
      if (res.statusCode !== 200) {
        reject(new Error(`HTTP ${res.statusCode} from ${url}`));
        return;
      }
      const chunks = [];
      res.on('data', (chunk) => chunks.push(chunk));
      res.on('end', () => resolve(Buffer.concat(chunks).toString()));
      res.on('error', reject);
    }).on('error', reject);
  });
}

function run(cmd, cwd = ROOT) {
  console.log(`  $ ${cmd}`);
  execSync(cmd, { cwd, stdio: 'inherit' });
}

function checkTool(name, installHint) {
  try {
    execSync(`which ${name}`, { stdio: 'pipe' });
  } catch {
    // Check in ~/go/bin for Go tools
    const goBinPath = path.join(process.env.HOME, 'go', 'bin', name);
    if (fs.existsSync(goBinPath)) return goBinPath;
    console.error(`\n  ✗ ${name} not found. Install with:\n    ${installHint}\n`);
    process.exit(1);
  }
  return name;
}

async function main() {
  const opts = parseArgs();

  // Step 1: Fetch spec
  if (!opts.skipFetch) {
    console.log(`\n→ Fetching OpenAPI spec from ${opts.url}`);
    try {
      const spec = await fetch(opts.url);
      // Validate it's valid JSON
      JSON.parse(spec);
      fs.writeFileSync(SPEC_PATH, spec);
      console.log(`  ✓ Saved to openapi/spec.json (${spec.length} bytes)`);
    } catch (err) {
      console.error(`  ✗ Failed to fetch spec: ${err.message}`);
      if (opts.url === URLS.local) {
        console.error('  Hint: is oway-platform-rest running locally on port 8181?');
      }
      process.exit(1);
    }
  } else {
    console.log('\n→ Skipping fetch, using existing openapi/spec.json');
  }

  // Step 2: TypeScript
  console.log('\n→ Generating TypeScript SDK');
  run('npx openapi-typescript ../../openapi/spec.json -o src/generated/schema.ts',
    path.join(ROOT, 'packages', 'typescript'));

  // Step 3: Go
  console.log('\n→ Generating Go SDK');
  const oapiCodegen = checkTool('oapi-codegen',
    `go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION}`);
  run(`${oapiCodegen} --config .oapi-codegen.yaml ../../openapi/spec.json`,
    path.join(ROOT, 'packages', 'go'));

  // Step 4: Build & test
  console.log('\n→ Building and testing Go SDK');
  run('go build ./...', path.join(ROOT, 'packages', 'go'));
  run('go test ./...', path.join(ROOT, 'packages', 'go'));

  console.log('\n→ Testing TypeScript SDK');
  run('npx vitest run', path.join(ROOT, 'packages', 'typescript'));

  console.log('\n✓ All SDKs regenerated and tests pass\n');
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
