// Metro config tuned for the pnpm + Turborepo monorepo.
//
// Metro needs to (1) watch the repo root so it picks up workspace packages
// such as @scrappd/shared-types, and (2) resolve modules from both the app's
// own node_modules and the hoisted root node_modules.
const { getDefaultConfig } = require('expo/metro-config');
const path = require('path');

const projectRoot = __dirname;
const monorepoRoot = path.resolve(projectRoot, '../..');

const config = getDefaultConfig(projectRoot);

config.watchFolders = [monorepoRoot];
config.resolver.nodeModulesPaths = [
  path.resolve(projectRoot, 'node_modules'),
  path.resolve(monorepoRoot, 'node_modules'),
];

module.exports = config;
