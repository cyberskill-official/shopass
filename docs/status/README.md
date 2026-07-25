# Status page

Generated artifact. Do not hand-edit `docs/status/index.html` or `docs/status/data/`.

## Source of truth

Task frontmatter (`status:` in each `docs/tasks/**/spec.md`) is the record of truth. This page is regenerated from that frontmatter by CyberOS:

```bash
bash .cyberos/lib/status-page.sh
```

## Drift note

If this page disagrees with task specs, regenerate with the command above. If the page and frontmatter agree but both disagree with reality (for example every task still marked `done` while code is missing), fix the frontmatter first (see hardening backlog item H1 / `docs/TASK-COVERAGE.md`), then regenerate.
