# FR Execution Done Checklist — SănDeal

Check every item before marking an FR task as `done` in tasks.md.

## 1. Read Docs
- [ ] `docs/feature-requests/SHIP-GUIDE.md`
- [ ] `docs/feature-requests/DATA-MODEL.md`
- [ ] The FR markdown file and companion `.audit.md`
- [ ] Relevant NFR files referenced by the FR
- [ ] FR frontmatter fields (`service`, `language`, `new_files`, `modified_files`, `allowed_tools`, `disallowed_tools`, `sub_tasks`)

## 2. Implement
- [ ] All files in `new_files` exist (or approved deviation recorded)
- [ ] Only `modified_files` were changed (no unrelated refactoring)
- [ ] Implementation follows §6 (Implementation Outline) of the FR
- [ ] All §1 MUST statements are implemented
- [ ] All §1 SHOULD statements are implemented or explicitly deferred
- [ ] SHIP-GUIDE invariants are preserved (no cleartext, BIGINT VND, MV3, user-initiated affiliate, etc.)
- [ ] Code follows the stack conventions from SHIP-GUIDE

## 3. Test §5
- [ ] All tests listed in §5 were run
- [ ] All tests pass
- [ ] If tests cannot be run, the blocker is documented with exact reason

## 4. Verify §1 -> §4
- [ ] Every §1 MUST/SHOULD statement maps to at least one §4 acceptance criterion
- [ ] Every §4 acceptance criterion is verifiable by a passing test or manual check
- [ ] Cross-reference recorded

## 5. Update Status
- [ ] FR frontmatter `status` field is updated according to STATUS-REFERENCE.md
- [ ] `docs/feature-requests/BACKLOG.md` status column matches FR frontmatter

## 6. Attach Evidence
- [ ] Evidence is recorded in the correct phase evidence file under `specs/001-full-project-plan/evidence/`
- [ ] Evidence includes: commit/ build ID, FR IDs, test results, migration version, security/compliance notes, known risks

## Blockers
If any item above cannot be completed, record the blocker:
- Task ID:
- Blocker description:
- FR(s) affected:
- Workaround or recommended action:
