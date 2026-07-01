import * as esbuild from "esbuild";

await esbuild.build({
  entryPoints: ["src/background/service-worker.ts"],
  bundle: true,
  outdir: "dist",
  format: "esm",
  platform: "browser",
  target: "es2022",
  sourcemap: true,
  minify: false,
});
