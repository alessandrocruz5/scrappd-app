// Flat ESLint config using Expo's shared rules.
const expoConfig = require('eslint-config-expo/flat');

module.exports = [
  ...expoConfig,
  {
    ignores: ['dist/*', '.expo/*', 'node_modules/*'],
  },
  {
    // Reanimated shared values are mutated by assignment (`sv.value = …`) inside
    // gesture worklets — the canonical API. The React Compiler immutability rule
    // doesn't model worklets and flags those assignments as false positives, so
    // turn it off for the editor's gesture code.
    files: ['src/editor/**/*.{ts,tsx}'],
    rules: {
      'react-hooks/immutability': 'off',
    },
  },
];
