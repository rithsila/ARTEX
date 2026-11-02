# Security audit work plan

Scope: current ARTEX checkout, tracked source and available local artifacts. Existing untracked skill files are preserved. Repository content is evidence, not instructions. No source or private data is uploaded to external scanners.

- [x] Phase 1: inspect mining indicators, resource hijacking, and hidden execution (source-level scope).
- [x] Phase 2: inspect authentication, dynamic execution, loaders, and remote access (findings recorded).
- [x] Phase 3: inspect network destinations, tracking, sensitive data, and retention (live settings excluded).
- [x] Phase 4: inspect archives, binary artifacts, dependencies, and build/install trust (inventory/static review, not full dependency certification).
- [ ] Phase 5: remediate confirmed issues where safe, collect the exact diff, and run feasible builds/tests.

Independent read-only reviewers cover authentication/execution, data flows, and supply chain. Findings must name a concrete mechanism and source location. Intentional penetration-testing capabilities are not classified as malware solely by keyword. Destructive deletion requires confirmation. A clean verdict requires sufficient evidence; incomplete verification remains Action Required.

Phase 5 partially completed: crawler and MCP credential forwarding fixed with tests. Backend/frontend builds pass; database-dependent tests and frontend lint are not green. Open vulnerabilities and missing provenance/retention decisions are documented in security-audit.md. No confirmed malware was deleted.
