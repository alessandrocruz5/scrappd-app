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
  collectCoverageFrom: ['src/**/*.{ts,tsx}', '!src/**/*.d.ts'],
  coverageReporters: ['text-summary', 'lcov'],
  // Coverage is reported, not gated. Uncomment to start enforcing a floor once
  // the suite is broad enough to make a threshold meaningful.
  // coverageThreshold: {
  //   global: { branches: 80, functions: 80, lines: 80, statements: 80 },
  // },
};
