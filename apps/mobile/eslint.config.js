// Flat ESLint config using Expo's shared rules.
const expoConfig = require('eslint-config-expo/flat');

module.exports = [
  ...expoConfig,
  {
    ignores: ['dist/*', '.expo/*', 'node_modules/*', 'coverage/*'],
  },
  {
    // Jest globals (describe/it/expect/jest) for the test suite so they aren't
    // flagged as undefined. @types/jest keeps the typecheck job happy too.
    files: ['**/*.test.{ts,tsx}', 'jest.setup.js'],
    languageOptions: {
      globals: {
        describe: 'readonly',
        it: 'readonly',
        test: 'readonly',
        expect: 'readonly',
        beforeAll: 'readonly',
        beforeEach: 'readonly',
        afterAll: 'readonly',
        afterEach: 'readonly',
        jest: 'readonly',
      },
    },
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
