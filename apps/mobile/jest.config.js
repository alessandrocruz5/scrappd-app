// Jest configuration for the Expo app. Uses the official `jest-expo` preset so
// React Native / Expo modules transform and resolve the way Metro expects.
module.exports = {
  preset: 'jest-expo',
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js'],
  // Mirror the `@/* -> ./src/*` alias from tsconfig.json. Jest does not read
  // TypeScript path mappings, so it has to be declared here too.
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
  },
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    '!src/**/*.d.ts',
    // Test scaffolding and the tests themselves aren't production code.
    '!src/**/__tests__/**',
    '!src/test-utils/**',
  ],
  coverageReporters: ['text-summary', 'lcov'],
  // A deliberately low but real floor. The data layer (cropper geometry +
  // cutout pipeline, books/pages API and its optimistic hooks, editor
  // transforms, page export) is well covered; the pure-presentation screens
  // that need a device renderer still pull the global down. This gate stops
  // regressions below today's baseline — raise the numbers as the component
  // suite grows.
  coverageThreshold: {
    global: {
      statements: 35,
      branches: 28,
      functions: 28,
      lines: 36,
    },
  },
};
