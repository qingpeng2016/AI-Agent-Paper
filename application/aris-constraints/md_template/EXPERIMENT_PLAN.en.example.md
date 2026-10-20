# Experiment Plan: [Short title aligned with rank-1 idea / FINAL_PROPOSAL]

Step 5 deliverable → `refine-logs/EXPERIMENT_PLAN.md`. Replace `[…]` with project-specific content from rank-1 idea, `FINAL_PROPOSAL.md`, and `RESEARCH_BRIEF.md`. **Do not ship finished plans with bracket placeholders.** Verify every official data URL in this run or mark `TBD — verify:`.

## Data access and workspace layout

### Dataset and official sources

[Dataset name and version for this study. Verified landing URL(s); separate download DOI/portal if applicable. DUA/license summary. Required citation.]

### How to obtain the data

[Ordered steps from verified instructions only. If not executed in this run, add subheading **Planned procedure (not yet executed)** and do not invent credentials or timings.]

### What you receive after download

[Directory tree, file types, approximate sizes, QC checklist — only after a real download or quoted from official docs with source link. Otherwise: **TBD until first download** and note what will be recorded in `data/<dataset-slug>/README.md`.]

### Where to put files in this project

[Paths relative to workspace root (contains `RESEARCH_BRIEF.md`), e.g. `data/<dataset-slug>/README.md`, `data/<dataset-slug>/raw/`, phenotypic/derivatives. No machine-specific absolute paths. No secrets in git.]

## Study question and claim hierarchy

**Question:** [One primary study question — claim-driven, testable with the locked validation design.]

**C1 (primary):** [Primary confirmatory claim, e.g. incremental predictive value of prespecified graph summaries over a locked FC baseline.]

**C2 (key secondary):** [Secondary claim on continuous clinical outcome or other prespecified endpoint.]

**C3 (robustness):** [Stability across held-out sites and prespecified construction sensitivities.]

**C4 (interpretive):** [What the study can and cannot conclude — e.g. predictive redundancy/complementarity, not mechanism or clinical utility by itself.]

## Phase 0 — Dataset and protocol lock

**Purpose:** Establish cohort and validation feasibility **without** outcome-guided feature or model selection.

1. Obtain the primary release under its terms; archive version, source URLs/access date, checksums, codebook, licensing.
2. Inspect **availability and QC only**: counts by site/group, key covariates, motion, scan length, derivatives. **Do not** fit diagnosis- or symptom-associated models for selection.
3. Apply locked eligibility rules; document exclusions and missingness.
4. Determine external holdout or nested leave-site-out feasibility; **freeze** holdout sites before outcome modeling. If infeasible, freeze repeated nested grouped CV as primary.
5. Register protocol, code skeleton, fold logic, feature/graph definitions, planned contrasts, exclusion report.

**Gate 0:** [Continue only if site-aware validation is viable and N supports stable grouped folds. Record decision and counts; amend design before outcomes if not. Phase 0 data steps **reference** “Data access and workspace layout” above — do not duplicate unverified URLs.]

## Phase 1 — Imaging QC and feature extraction

1. [Lock atlas/parcellation and network labels.]
2. [Locked preprocessing/QC; motion and retained scan lengths by site and group.]
3. [FC construction rules: edge definition, diagonal, missing parcels.]
4. [FC feature families: e.g. within-network means, network-pair means, matched-dimensionality PCA baseline.]
5. [Graph construction at prespecified densities; graph metrics with fixed definitions and seeds; density AUC summary if used.]
6. [Prespecified sensitivity features — binary/weighted/GSR etc. — without outcome-guided selection.]
7. [Sanity checks on synthetic or held-aside matrices; separate from outcome inference.]

**Deliverables:** [Data-flow diagram/table, feature dictionary, QC report, versioned code, immutable participant-to-feature mapping.]

## Phase 2 — Nested site-aware model fitting

### Diagnosis (or primary classification outcome)

[Identical nuisance structure across compared models; prespecified model ladder from nuisance-only through baseline FC block, graph block, combined, dimensionality-matched sensitivity. Same outer folds for all comparisons.]

### Continuous symptom / secondary outcome

[Corresponding linear (or prespecified) models on the eligible subset; complete-case policy; disclose N and missingness. No post hoc dichotomization unless preregistered.]

### Fold and leakage policy

[Site holdout + grouped nested CV on discovery sites, or repeated nested CV with sites grouped in outer/inner folds. All imputation, scaling, PCA, harmonization, regularization, and threshold tuning **inside training folds only**. Store out-of-fold predictions before aggregation.]

### Metrics

[Primary: paired ΔAUC / Δlog loss / calibration for classification; ΔR² / ΔRMSE for regression. Intervals via site-aware bootstrap or grouped permutation; state resampling unit.]

## Phase 3 — Confirmatory inference

1. [Test C1 on locked validation predictions — estimate, interval, calibration, site heterogeneity.]
2. [Fixed-sequence gate for C2 if prespecified; report C2 regardless of gate outcome.]
3. [Dimensionality-matched baseline comparison — interpret concordance or reversal.]
4. [Per-site and leave-one-site-out contrasts; do not claim transportability from pooled CV alone.]
5. [Confirmatory multiplicity procedure; FDR within exploratory families with all estimates visible.]

## Phase 4 — Robustness and failure analysis

[Run all prespecified sensitivities irrespective of primary result: graph construction, density, preprocessing, nuisance adjustment, ablations, baselines, site-stratified vs leave-site-out. Multiverse table with direction, magnitude, intervals.]

[Null-over-null: absolute performance, calibration, uncertainty, site heterogeneity; bounded claims if both baseline and combined models are weak.]

## Phase 5 — Reporting and release

[Manuscript structure tied to C1–C4; flow diagram, site table, feature definitions, validation schematic, performance plots, site increment plot, sensitivity matrix. Protocol deviations and all planned analyses. Code/environment/folds/aggregates per DUA. Results-to-claim table to avoid mechanistic overreach.]

## Required figures and tables

1. [Cohort flow and site/group counts.]
2. [Feature construction diagram — baseline vs graph blocks.]
3. [Nested / site-aware validation schematic.]
4. [Out-of-sample ROC/calibration and paired Δ metric intervals.]
5. [Secondary outcome observed/predicted and paired intervals.]
6. [Per-site / leave-one-site-out incremental effect plot.]
7. [Sensitivity matrix for all prespecified construction choices.]

## Preregistered decision table

| Observed pattern | Permitted conclusion |
|---|---|
| [Primary increment positive, interval excludes null, site-consistent, dimensionality control concordant] | [Claim-level conclusion for C1.] |
| [Positive point estimate but wide interval or site heterogeneity] | [Uncertainty / heterogeneity wording.] |
| [Main positive but matched-dimensionality baseline null/reversed] | [No general increment claim beyond representation choice.] |
| [No graph increment; FC baseline predictive] | [Limited added value of graph block under tested conditions.] |
| [Both weak] | [Limited predictive utility in this cohort; not proof of biological absence.] |
| [Direction flips across construction/preprocessing] | [Analytic-choice dependence.] |
| [Classification increment without symptom increment] | [Outcome-specific interpretation.] |

## Milestones and dependencies

| Milestone | Output | Dependency / stop condition |
|---|---|---|
| M0: Data lock and feasibility | [Versioned cohort spec and site map] | [Stop/revise if grouped validation invalid or critical fields missing.] |
| M1: Protocol freeze | [Timestamped plan and code skeleton] | [Before any outcome modeling.] |
| M2: QC and features | [QC report and frozen feature tables] | [No outcome inspection for feature fixes.] |
| M3: Nested fitting | [Fold-wise predictions and logs] | [All transforms fold-local.] |
| M4: Confirmatory analysis | [C1/C2 estimates, intervals, multiplicity record] | [Fixed sequence; report all prespecified tests.] |
| M5: Robustness and reporting | [Sensitivity matrix, figures, manuscript draft] | [All variants and deviations included.] |
