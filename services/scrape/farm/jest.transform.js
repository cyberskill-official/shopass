// Minimal Jest transformer: transpile TypeScript to CommonJS with the installed
// `@typescript/typescript6` compiler (TS6 JS API; project typechecks with typescript@7 CLI). Type-checking is done separately by `tsc --noEmit`;
// this only strips types so Jest can run the .test.ts files. Keeps emit on the TS6 JS API,
// which avoids the ts-jest peer range (<7) and TS7 lacking transpileModule.
const ts = require('@typescript/typescript6');

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
