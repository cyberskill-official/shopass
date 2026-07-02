// Minimal Jest transformer: transpile TypeScript to CommonJS with the installed
// `typescript` compiler. Type-checking is done separately by `tsc --noEmit`;
// this only strips types so Jest can run the .test.ts files. Zero extra deps,
// which avoids the ts-jest/jest-30/TS-6 resolution breakage.
const ts = require('typescript');

module.exports = {
  process(sourceText, sourcePath) {
    const out = ts.transpileModule(sourceText, {
      compilerOptions: {
        module: ts.ModuleKind.CommonJS,
        target: ts.ScriptTarget.ES2020,
        esModuleInterop: true,
        sourceMap: true,
        inlineSources: true,
      },
      fileName: sourcePath,
    });
    return { code: out.outputText, map: out.sourceMapText };
  },
};
