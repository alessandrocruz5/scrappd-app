// Metro config tuned for the pnpm + Turborepo monorepo.
//
// Metro needs to (1) watch the repo root so it picks up workspace packages
// such as @scrappd/shared-types, and (2) resolve modules from both the app's
// own node_modules and the hoisted root node_modules.
//
// We start from Sentry's `getSentryExpoConfig` instead of Expo's
// `getDefaultConfig` so production builds emit and (with SENTRY_AUTH_TOKEN set)
// upload source maps — turning minified web/native stack frames back into
// readable file/line references in the Sentry dashboard. It's a drop-in
// superset of the Expo default config, so the monorepo tweaks below still
// apply.
const { getSentryExpoConfig } = require('@sentry/react-native/metro');
const path = require('path');

const projectRoot = __dirname;
const monorepoRoot = path.resolve(projectRoot, '../..');

const config = getSentryExpoConfig(projectRoot);

config.watchFolders = [monorepoRoot];
config.resolver.nodeModulesPaths = [
  path.resolve(projectRoot, 'node_modules'),
  path.resolve(monorepoRoot, 'node_modules'),
];

module.exports = config;
