import * as fs from "fs";
import * as path from "path";

interface HookResult {
  name: string;
  pass: boolean;
  detail?: string;
  [key: string]: unknown;
}

interface AuditReport {
  version: string;
  commit: string;
  verdict: "PASS" | "FAIL";
  hooks: HookResult[];
}

const ROOT = path.resolve(__dirname, "..");
const RESULT_DIR = path.join(ROOT, "report", "hook-results");
const REQUIRED_HOOKS = ["egress", "sbom", "reproducible", "payload_guard"] as const;

function loadHook(name: string): HookResult {
  const file = path.join(RESULT_DIR, `${name}.json`);
  if (!fs.existsSync(file)) {
    return {
      name,
      pass: false,
      detail: "NOT_RUN — no hook-results file (fail-closed)",
    };
  }
  try {
    const parsed = JSON.parse(fs.readFileSync(file, "utf8")) as HookResult;
    const detail =
      typeof parsed.detail === "string" ? parsed.detail : undefined;
    return {
      ...parsed,
      name,
      pass: Boolean(parsed.pass),
      detail,
    };
  } catch (err) {
    return {
      name,
      pass: false,
      detail: `NOT_RUN — failed to parse hook result: ${String(err)}`,
    };
  }
}

const hooks: HookResult[] = REQUIRED_HOOKS.map(loadHook);
const pass = hooks.length === REQUIRED_HOOKS.length && hooks.every((h) => h.pass);

const report: AuditReport = {
  version: "1.4.0",
  commit: process.env.GITHUB_SHA || "local",
  verdict: pass ? "PASS" : "FAIL",
  hooks,
};

fs.mkdirSync(path.join(ROOT, "report"), { recursive: true });
fs.writeFileSync(
  path.join(ROOT, "report", "audit-report.json"),
  JSON.stringify(report, null, 2) + "\n",
);

const statusCell = (h: HookResult) => (h.pass ? "PASS" : "FAIL / NOT_RUN");
const evidence = (h: HookResult) => h.detail ?? "(no detail)";

const md = `# SănDeal Security Audit

Version: ${report.version}
Commit: ${report.commit}

| Hook | Kết quả | Bằng chứng |
|---|---|---|
| Egress (động) | ${statusCell(hooks[0])} | ${evidence(hooks[0])} |
| SBOM + vuln | ${statusCell(hooks[1])} | ${evidence(hooks[1])} |
| Reproducible build | ${statusCell(hooks[2])} | ${evidence(hooks[2])} |
| Payload guard (tĩnh) | ${statusCell(hooks[3])} | ${evidence(hooks[3])} |

=> AUDIT ${report.verdict}. Hooks that are not implemented fail closed. Re-run: \`bash audit/run-security-audit.sh\`
`;

fs.writeFileSync(path.join(ROOT, "report", "audit-report.md"), md);
console.log(`Report generated: AUDIT ${report.verdict}`);
if (!pass) {
  process.exitCode = 1;
}
