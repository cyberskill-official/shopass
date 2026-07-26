module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
  testMatch: ["**/__tests__/**/*.test.ts", "**/src/**/*.test.ts"],
  moduleNameMapper: {
    "^react-native-keychain$": "<rootDir>/src/auth/__mocks__/react-native-keychain.ts",
  },
};
