# Experiment Plan: [Short title from rank-1 / FINAL_PROPOSAL]

**Template for standalone `/experiment-plan`** (not Step 5 `research-refine-pipeline`).  
Fill from `experiment-plan.manifest.example.desc`, [ARIS `experiment-plan/SKILL.md`](https://github.com/wanshuiyin/Auto-claude-code-research-in-sleep/blob/main/skills/experiment-plan/SKILL.md), `RESEARCH_BRIEF.md`, `refine-logs/FINAL_PROPOSAL.md` (if present), and rank-1 in `idea-stage/IDEA_REPORT.en.md`.  
**Output:** `refine-logs/EXPERIMENT_PLAN.en.md`. Pair with `EXPERIMENT_PLAN.zh.example.md` (same blocks; Chinese prose).  
**Do not leave bracket placeholders in finished deliverables.**

**Problem:** [One paragraph — reproducibility bottleneck / question]  
**Method Thesis:** [One sentence — dominant method or analysis framework]  
**Date:** [YYYY-MM-DD]

---

## Claim Map

**Note (optional in deliverables):** **C** = claim, **E** = experiment; **C1** = primary claim (≥1 row), **C2** = supporting claim (optional). **Linked Blocks:** E1, E2, …

| Claim | Why It Matters | Minimum Convincing Evidence | Linked Blocks |
|-------|----------------|----------------------------|---------------|
| **C1:** [Primary claim] | [Why a reviewer should care] | [Smallest evidence that would convince] | [E1, E2, …] |
| **C2:** [Optional supporting claim] | … | … | … |

---

## Paper Storyline

### Main paper must prove:
1. [Claim / result line 1]
2. [Claim / result line 2]

### Appendix can support:
- [Sensitivity, extra figures, optional stratifications]

### Experiments intentionally cut:
- [Scope exclusions — not part of this paper]

---

## Experiment Data Acquisition

### Experimental acquisition

For **data you generate or coordinate** (not only public downloads). Specify subjects, protocol, and on-disk layout.

#### Subjects or units of study

[Cohort, n, inclusion/exclusion, groups, sites/instruments, ethics/consent if applicable]

#### Methods

[Protocol, acquisition pipeline, QC, fields logged; link to experiment blocks E1…]

#### Data products and layout

[Output layout/schema; mark **schema verified** (protocol doc or sample) with date; no unverified fields]

### Data download

For **externally hosted** datasets or benchmarks.

#### Download URLs

[**Prefer free direct-download** sources (open mirrors, public buckets, Zenodo/OpenNeuro, etc.) before gated/paid; verify reachable/downloadable; **verification status** + date]

#### Access procedure

[For direct links: note no extra steps; for gated access: registration/DUA/purchase steps; **Planned procedure (not yet executed)** if applicable]

#### Data products and layout

[Post-download layout; mapping to `data/<slug>/`; **schema verified** against official docs or sample file; QC/checksums]

---

## Experiment Blocks

**Ground every block in `Experiment Data Acquisition` above**—same paths, cohort, and schema fields; optional **Data dependency:** method 1/2 + path.

### Block E1: [Name]

**Claim tested:** [C1 / C2]

**Why this block exists:** [One short paragraph]

**Data dependency:** [Method 2 → `data/…`; or Method 1 output; columns match acquisition **Data products and layout**]

**Dataset / split / task:**
- [Cohort, n, sites, task]

**Compared systems:**
1. [Primary method / multiverse / main pipeline]
2. [Strong baselines — ≤3 families per ARIS limits]
3. [Optional ablation or negative control]

**Metrics** (primary):
- [Decisive metrics first]

**Setup details:**
- [Model, covariates, seeds, multiplicity, compute order-of-magnitude]

**Success criterion:**
- [Pass/fail thresholds — pre-specified]

**Failure interpretation:**
- [What a negative result means]

**Table / figure target:**
- [Main paper vs appendix targets]

**Priority:** **MUST-RUN** | **NICE-TO-HAVE**

---

### Block E2: [Name]

[Repeat E1 field set for each block — ≤5 core blocks]

---

## Run Order and Milestones

| Milestone | Goal | Runs | Decision Gate | GPU-hrs | Wall-clock | Risk | Mitigation |
|-----------|------|------|---------------|---------|------------|------|------------|
| **M0** | [Sanity / data QC] | R001–… | [Gate] | … | … | … | … |
| **M1** | [Core evidence] | … | … | … | … | … | … |

**Total estimated GPU-hours:** [number]  
**Timeline:** [weeks or phases]  
**Critical path:** [M0 → …]

---

## Compute and Data Budget

**Compute:** GPU-hours, CPU, memory, storage  
**Data preparation:** access, harmonization, exclusions  
**Bottleneck analysis:** [dominant constraint]

---

## Risks and Mitigations

| Risk | Severity | Consequence | Mitigation |
|------|----------|-------------|------------|
| [Risk 1] | … | … | … |

---

## Final Checklist

- [ ] Main paper tables/figures covered by must-run blocks
- [ ] Novelty / mechanism isolated (not only aggregate performance)
- [ ] Simplicity defended (no unnecessary complexity)
- [ ] Must-run vs nice-to-have separated
- [ ] Preregistration-ready (metrics, thresholds, models frozen)
- [ ] Download URLs prefer free direct access (or justify gated-only source)
- [ ] Download URLs verified reachable/downloadable (or honestly marked credential-gated)
- [ ] Data layouts match actual files when present (not doc-only fantasy)
- [ ] Every experiment block cites acquisition data only (no orphan datasets)
- [ ] Data access and compute budget realistic

**Optional:** `## Key Decisions Frozen` — numbered list of locked design choices (if helpful for your study).
