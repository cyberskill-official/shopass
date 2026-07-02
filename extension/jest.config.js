/** @type {import('jest').Config} */
module.exports = {
  testEnvironment: 'jsdom',
  transform: { '^.+\\.(t|j)sx?$': 'babel-jest' },
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node'],
  testMatch: ['**/test/**/*.test.ts', '**/src/**/*.test.ts'],
};
