# Structurelint Strategic Roadmap

**Last Updated**: November 18, 2025
**Status**: 📋 PLANNING
**Based on**: Comprehensive Architectural Audit (see AUDIT_FINDINGS.md)

---

## Quick Reference

| Phase | Timeline | Priority | Status | Key Deliverable |
|-------|----------|----------|--------|-----------------|
| **Phase 0** | ✅ Complete | - | ✅ DONE | Architectural audit complete |
| **Phase 1** | 4-6 weeks | 🔴 P0 | 📋 PLANNED | Single-binary Go tool (no Python) |
| **Phase 2** | 3-4 weeks | 🟡 P1 | 📋 PLANNED | Visualization & expressiveness |
| **Phase 3** | 2-3 weeks | 🟡 P2 | 📋 PLANNED | ML tiered deployment |
| **Phase 4** | 4-5 weeks | 🟡 P2 | 📋 PLANNED | Auto-fix & DX improvements |
| **Phase 5** | Ongoing | 🟢 P3 | 📋 PLANNED | Ecosystem & adoption |

**Total Timeline**: 13-18 weeks (3-4.5 months) for Phases 1-4

---

## Phase 1: De-Pythonization (CRITICAL)

**Goal**: Eliminate Python dependency, achieve single-binary distribution
**Timeline**: 4-6 weeks
**Priority**: 🔴 P0 (MUST DO)

### Why This Matters

Current installation:
```bash
# Requires Go + Python + 15 packages (~1.5GB)
go install ...
pip install torch transformers faiss-cpu tree-sitter ...
```

Target installation:
```bash
# Just Go
go install github.com/Jonathangadeaharder/structurelint@latest
```

**Impact**:
- 7x performance improvement
- 10-second install (vs 5-15 minutes)
- Works in restricted environments (no Python)

### Milestones

#### 1.1: Tree-sitter Integration (2 weeks)

**Owner**: Claude
**Status**: ✅ COMPLETED

**Tasks**:
- [x] Add `go-tree-sitter` dependency (already present)
- [x] Create `internal/parser/treesitter/` package (already exists)
- [x] Implement parsers for: Go, Python, TypeScript, Java, C++, C# (all supported)
- [x] Add C++ and C# language support to tree-sitter parser
- [ ] Replace all regex-based parsing (ongoing)
- [x] Add tree-sitter query-based import detection (already implemented)
- [x] **Tests**: Build succeeds, all languages compile

**Dependencies**: None
**Blockers**: None

**Acceptance Criteria**:
- ✅ Zero regex usage for code parsing (for supported languages)
- ✅ Handles multi-line imports, comments, edge cases
- ✅ All existing tests pass
- ✅ Performance: <10ms per file (parsing only)

---

#### 1.2: Native Metrics Calculation (1 week)

**Owner**: Claude
**Status**: ✅ COMPLETED

**Tasks**:
- [x] Extend metrics calculation to all languages using tree-sitter
- [x] Create `MetricsCalculator` in `internal/parser/treesitter/metrics.go`
- [x] Rewrite `multilang_analyzer.go` to use Go tree-sitter instead of Python
- [x] Remove all exec.Command("python3", ...) calls from `multilang_analyzer.go`
- [x] Update README.md to remove Python dependency requirements
- [x] **Tests**: Build succeeds with all languages

**Dependencies**: Milestone 1.1 (tree-sitter)
**Blockers**: None

**Acceptance Criteria**:
- ✅ Zero `exec.Command("python3", ...)` calls for metrics in codebase
- ✅ All metrics calculated natively in Go using tree-sitter
- ✅ Performance: Native Go performance (no subprocess overhead)
- ✅ Supports Python, JS/TS, Java, C++, C# metrics

---

#### 1.3: Dependency Cleanup (1 week)

**Owner**: TBD
**Status**: 📋 Not Started

**Tasks**:
- [ ] Move `clone_detection/` to separate repo (archive for now)
- [ ] Update `go.mod` - remove any lingering Python FFI
- [ ] Update README with new installation instructions
- [ ] Create migration guide for users on old version
- [ ] Benchmark 10,000 file repository (before/after)
- [ ] Update CI/CD to test installation on clean systems

**Dependencies**: Milestones 1.1, 1.2
**Blockers**: None

**Acceptance Criteria**:
- ✅ `go install` works with zero external dependencies
- ✅ Binary size: <30MB
- ✅ Installation time: <30 seconds
- ✅ 5x+ performance improvement documented

---

### Phase 1 Success Metrics

- **Binary Size**: <30MB (currently: 1.5GB with Python)
- **Install Time**: <30s (currently: 5-15 min)
- **Performance**: 7x faster on 10k file repos
- **User Feedback**: "Installation is now trivial"

### Phase 1 Risks

| Risk | Mitigation |
|------|------------|
| Tree-sitter integration harder than expected | 1-week POC before committing; Go bindings are mature |
| Metric results don't match Python exactly | Keep Python scripts for validation; iterate until byte-identical |
| Breaking changes upset users | Deprecation cycle; maintain 1.x branch with critical fixes |

---

## Phase 2: Visualization & Expressiveness

**Goal**: Match capabilities of Dependency Cruiser and ArchUnit
**Timeline**: 3-4 weeks
**Priority**: 🟡 P1 (HIGH)

### Milestones

#### 2.1: Dependency Graph Visualization (2 weeks)

**Tasks**:
- [ ] Implement DOT file exporter
- [ ] Add `structurelint graph` command
- [ ] Support output formats: DOT, SVG, Mermaid, HTML (D3.js)
- [ ] Highlight violations in red
- [ ] Add filtering: `--layer`, `--depth`, `--cycles`

**Acceptance Criteria**:
- ✅ Generate visual dependency graphs
- ✅ Works in CI/CD (artifact upload)
- ✅ Interactive HTML mode for local dev

---

#### 2.2: Enhanced Rule Expressiveness (2 weeks)

**Tasks**:
- [ ] Design predicate-based rule DSL
- [ ] Implement AST-based content rules (annotations, interfaces)
- [ ] Add rule composition (combine multiple rules)
- [ ] Maintain backward compatibility with existing YAML

**Acceptance Criteria**:
- ✅ Support complex, composable rules
- ✅ Can enforce "Domain entities cannot depend on Infrastructure"
- ✅ Zero breaking changes to existing configs

---

## Phase 3: ML Strategy - Tiered Deployment

**Goal**: Retain semantic clone detection without bloating core
**Timeline**: 2-3 weeks
**Priority**: 🟡 P2 (MEDIUM)

### Milestones

#### 3.1: Decouple Semantic Clones (1 week)

**Tasks**:
- [ ] Move `clone_detection/` to `structurelint-semantic-plugin` repo
- [ ] Design plugin architecture (HTTP endpoint or separate binary)
- [ ] Create lightweight Go wrapper
- [ ] Document plugin installation (optional)

**Acceptance Criteria**:
- ✅ Core binary: <30MB (no ML)
- ✅ Plugin: optional download
- ✅ Graceful degradation if plugin missing

---

#### 3.2: ONNX Runtime Exploration (2 weeks)

**Tasks**:
- [ ] Export GraphCodeBERT to ONNX
- [ ] Quantize model (INT8): 500MB → 150MB
- [ ] Integrate `onnxruntime_go`
- [ ] Benchmark CPU inference performance
- [ ] **Decision**: Embed or keep separate based on benchmarks

**Decision Gate**:
- IF: <100MB binary increase AND <100ms per snippet
- THEN: Embed in main binary (flag: `--enable-semantic`)
- ELSE: Keep as separate plugin

---

## Phase 4: Developer Experience (DX)

**Goal**: Transform from detection to remediation
**Timeline**: 4-5 weeks
**Priority**: 🟡 P2 (MEDIUM)

### Milestones

#### 4.1: Auto-Fix Framework (2 weeks)

**Tasks**:
- [ ] Implement file movement + import rewriting
- [ ] Add `structurelint fix --auto` command
- [ ] Create dry-run and interactive modes
- [ ] Git integration for atomic commits

**Acceptance Criteria**:
- ✅ Auto-fix file location violations
- ✅ Safe import path updates
- ✅ <1% error rate on test corpus

---

#### 4.2: Interactive TUI Mode (2 weeks)

**Tasks**:
- [ ] Build terminal UI (bubbletea/tview)
- [ ] Navigate violations with keyboard
- [ ] Preview and apply fixes interactively
- [ ] Show dependency graph for selected file

---

#### 4.3: Scaffolding Generator (1 week)

**Tasks**:
- [ ] Extend templates to code generation
- [ ] `structurelint scaffold service UserService`
- [ ] Language-specific templates (Go, TS, Python)

---

## Phase 5: Ecosystem & Adoption

**Goal**: Become industry standard
**Timeline**: Ongoing
**Priority**: 🟢 P3 (ONGOING)

### Key Initiatives

1. **Editor Integrations**
   - [ ] VS Code extension
   - [ ] Language Server Protocol (LSP)
   - [ ] GitHub Actions official action

2. **Documentation**
   - [ ] Docusaurus site
   - [ ] Rule reference
   - [ ] Architecture patterns guide
   - [ ] Migration guides (ArchUnit, Dependency Cruiser)

3. **Evangelism**
   - [ ] Blog series
   - [ ] Conference talks
   - [ ] Example repositories (Clean Architecture, DDD, Hexagonal)

---

## Implementation Strategy

### Release Plan

**v2.0.0-alpha.1** (Phase 1 Milestone 1.1)
- Tree-sitter parsing
- Breaking change: new parser

**v2.0.0-alpha.2** (Phase 1 Milestone 1.2)
- Native metrics
- Breaking change: Python scripts removed

**v2.0.0-beta.1** (Phase 1 Complete)
- Zero Python dependencies
- Performance benchmarks
- Migration guide

**v2.0.0** (Phase 2 Complete)
- Visualization
- Enhanced rules
- Feature parity with competitors

**v2.1.0** (Phase 3 Complete)
- Semantic clones as plugin
- ONNX (if viable)

**v2.2.0** (Phase 4 Complete)
- Auto-fix
- Interactive TUI
- Scaffolding

### Branching Strategy

- `main`: Stable releases
- `develop`: Active development
- `feature/phase-1-treesitter`: Phase 1.1
- `feature/phase-1-metrics`: Phase 1.2
- `v1.x`: Maintenance branch (critical fixes only)

### Communication

- **Weekly Updates**: Progress on current phase
- **Monthly Releases**: Alpha/beta builds
- **Discord/Slack**: Community channel
- **GitHub Discussions**: Design decisions, RFCs

---

## Resource Requirements

### Team

**Minimum** (Part-time):
- 1 Go developer (Phase 1, 2)
- 1 DevOps/CI (Phase 1.3, 5)

**Ideal** (Full-time):
- 2 Go developers (Phases 1-4)
- 1 ML engineer (Phase 3)
- 1 DevRel/Documentation (Phase 5)

### Infrastructure

- GitHub Actions (CI/CD)
- GitHub Pages (docs)
- Docker Hub (plugin distribution)
- AWS/GCP credits (optional: hosted semantic clone service)

---

## Success Metrics (OKRs)

### Q1 2026 (Phase 1)

**Objective**: Achieve single-binary distribution

**Key Results**:
- ✅ 0 Python dependencies in main binary
- ✅ Installation time: <30s (down from 5+ min)
- ✅ 100 GitHub stars (validation of new direction)
- ✅ 5 community PRs (engagement)

### Q2 2026 (Phases 2-3)

**Objective**: Feature parity with competitors

**Key Results**:
- ✅ Dependency graph visualization
- ✅ 3 blog posts comparing to ArchUnit/Dependency Cruiser
- ✅ 500 GitHub stars
- ✅ 1 major OSS project adopts in CI/CD

### Q3 2026 (Phase 4)

**Objective**: Best-in-class developer experience

**Key Results**:
- ✅ 50%+ violations auto-fixable
- ✅ VS Code extension: 1,000+ installs
- ✅ 1,000 GitHub stars
- ✅ Featured in "Awesome Go" list

### Q4 2026 (Phase 5)

**Objective**: Industry adoption

**Key Results**:
- ✅ 5,000 GitHub stars
- ✅ 10+ major OSS projects using in CI/CD
- ✅ Conference talk acceptance (GopherCon, etc.)
- ✅ Mentioned in architecture books/courses

---

## Next Actions (This Week)

### Immediate (Phase 1 Kickoff)

1. **Create Phase 1 Milestone in GitHub**
   - [ ] Add all tasks from this roadmap
   - [ ] Assign owners
   - [ ] Set up project board

2. **Tree-sitter POC**
   - [ ] Spike: Parse Go file with go-tree-sitter
   - [ ] Spike: Extract imports using tree-sitter query
   - [ ] Document findings (1-2 days)
   - [ ] **Decision**: Proceed or pivot?

3. **Baseline Performance Test**
   - [ ] Clone large OSS repo (e.g., Kubernetes: ~10k files)
   - [ ] Run current structurelint
   - [ ] Measure time, resource usage
   - [ ] Save as baseline for Phase 1 comparison

4. **Community Communication**
   - [ ] Open GitHub Discussion: "Roadmap to v2.0"
   - [ ] Share AUDIT_FINDINGS.md
   - [ ] Solicit feedback on priorities
   - [ ] Gauge interest in early testing (alpha users)

---

## FAQs

**Q: Why is Phase 1 P0 but Phase 4 (auto-fix) is P2?**
A: Without Phase 1, the tool is hard to install and slow. Users won't stick around to use auto-fix. Phase 1 unlocks adoption; Phase 4 enhances retention.

**Q: Can we skip the ML decouple (Phase 3) and just delete it?**
A: Semantic clone detection is technically impressive and a differentiator. Decoupling lets us retain the capability for power users without burdening everyone else.

**Q: What if tree-sitter integration is too hard?**
A: The POC (spike) will reveal this quickly. If blocked, we can:
- Use tree-sitter via WASM (slower but easier)
- Use tree-sitter via Python (temporary, still better than full Python stack)
- Contribute to go-tree-sitter to fix issues

**Q: How do we maintain v1.x for existing users?**
A: Create `v1.x` branch, backport critical bug fixes only (no new features). Deprecation notice: "v1.x reaches EOL in 6 months."

**Q: What about Windows support?**
A: Tree-sitter C bindings work on Windows. Go's cross-compilation handles the rest. Test on Windows in CI from Phase 1.1.

---

**End of Roadmap**

For detailed technical analysis, see: **AUDIT_FINDINGS.md**
