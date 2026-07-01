import * as esbuild from 'esbuild';

const isWatch = process.argv.includes('--watch');

process.env.SOURCE_DATE_EPOCH = process.env.SOURCE_DATE_EPOCH || "0";

const ctx = await esbuild.context({
  entryPoints: ['src/background/service-worker.ts', 'src/content/shopee/index.ts'],
  bundle: true,
  outdir: 'dist',
  format: 'esm',
  target: ['es2022'],
  minify: true,
  absWorkingDir: process.cwd(),
  sourcemap: false,
  legalComments: 'none',
  define: { "process.env.BUILD_TIME": '""' },
  metafile: false,
});

if (isWatch) {
  await ctx.watch();
  console.log('Watching for changes...');
} else {
  await ctx.rebuild();
  await ctx.dispose();
  console.log('Build complete.');
}
