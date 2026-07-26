import * as esbuild from 'esbuild';
import * as fs from 'fs';
import * as path from 'path';

const isWatch = process.argv.includes('--watch');

process.env.SOURCE_DATE_EPOCH = process.env.SOURCE_DATE_EPOCH || "0";

const shopassEnv = process.env.SHOPASS_ENV || "production";
if (!["development", "staging", "production"].includes(shopassEnv)) {
  console.error(`Invalid SHOPASS_ENV=${shopassEnv} (use development|staging|production)`);
  process.exit(1);
}

const entryPoints = [
  'src/background/service-worker.ts',
  'src/content/shopee/index.ts',
  'src/content/tiktok/index.ts',
  'src/content/lazada/index.ts',
  'src/ui/settings.ts',
  'src/ui/onboarding.ts'
];

const ctx = await esbuild.context({
  entryPoints,
  bundle: true,
  outdir: 'dist',
  format: 'esm',
  target: ['es2022'],
  minify: true,
  absWorkingDir: process.cwd(),
  sourcemap: false,
  legalComments: 'none',
  define: {
    "process.env.BUILD_TIME": '""',
    "process.env.SHOPASS_ENV": JSON.stringify(shopassEnv),
  },
  metafile: false,
});

function copyAssets() {
  // Ensure directories exist
  fs.mkdirSync('dist/ui', { recursive: true });

  // Copy static assets including audited DNR allowlist rules (R31).
  fs.copyFileSync('manifest.json', 'dist/manifest.json');
  fs.copyFileSync('src/ui/settings.html', 'dist/ui/settings.html');
  fs.copyFileSync('src/ui/onboarding.html', 'dist/ui/onboarding.html');
  fs.mkdirSync('dist/dnr', { recursive: true });
  fs.copyFileSync('src/dnr/rules.json', 'dist/dnr/rules.json');
  console.log('Static assets copied to dist.');
}

if (isWatch) {
  await ctx.watch();
  console.log('Watching for changes...');
} else {
  await ctx.rebuild();
  await ctx.dispose();
  copyAssets();
  console.log(`Build complete (SHOPASS_ENV=${shopassEnv}).`);
}
