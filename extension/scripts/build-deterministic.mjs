import * as esbuild from "esbuild";

process.env.SOURCE_DATE_EPOCH = process.env.SOURCE_DATE_EPOCH || "0";

await esbuild.build({
  entryPoints: ["src/background/service-worker.ts", "src/content/shopee/index.ts"],
  bundle: true,
  outdir: "dist",
  format: "esm",
  target: ["es2022"],
  minify: true,
  absWorkingDir: process.cwd(),
  sourcemap: false,
  legalComments: "none",
  define: { "process.env.BUILD_TIME": '""' },
  metafile: false,
});
