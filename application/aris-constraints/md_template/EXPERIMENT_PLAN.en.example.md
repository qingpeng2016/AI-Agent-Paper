# Experiment Plan: [Short title from rank-1 idea / FINAL_PROPOSAL]

Template for Step 5 (rank-1 only). Fill all sections from `research-refine-pipeline.manifest.example.desc`, rank-1 content in `idea-stage/IDEA_REPORT.en.md` (§3 Idea 1, §4–§5 for that idea), `FINAL_PROPOSAL.md`, and `RESEARCH_BRIEF.md`. Output: **`refine-logs/EXPERIMENT_PLAN.md`**. Pair with `EXPERIMENT_PLAN.zh.example.md` (same `sec-*` anchors). **Do not leave bracket placeholders in finished sections** — replace with real prose, lists, and tables suited to **this** study (any field).

## Table of contents

- [1. Data access and workspace layout](#sec-1)
  - [1.1 Dataset and official sources](#sec-1-1)
  - [1.2 How to obtain the data](#sec-1-2)
  - [1.3 What you receive after download](#sec-1-3)
  - [1.4 Where to put files in this project](#sec-1-4)
- [2. Study question and claim hierarchy](#sec-2)
  - [2.1 Primary question](#sec-2-1)
  - [2.2 Confirmatory claims (fixed order)](#sec-2-2)
  - [2.3 Interpretive bounds](#sec-2-3)
- [3. Phase 0 — Data and protocol lock](#sec-3)
- [4. Phase 1 — Preparation and analysis-ready inputs](#sec-4)
- [5. Phase 2 — Primary analysis pipeline](#sec-5)
  - [5.1 Primary endpoint / main test](#sec-5-1)
  - [5.2 Secondary and exploratory analyses](#sec-5-2)
  - [5.3 Validation design and leakage controls](#sec-5-3)
  - [5.4 Metrics, intervals, and multiplicity](#sec-5-4)
- [6. Phase 3 — Confirmatory inference](#sec-6)
- [7. Phase 4 — Robustness and sensitivity](#sec-7)
- [8. Phase 5 — Reporting and release](#sec-8)
- [9. Required figures and tables](#sec-9)
- [10. Preregistered decision table](#sec-10)
- [11. Milestones and dependencies](#sec-11)

---

## Data access and workspace layout

<a id="sec-1"></a>

### 1. Data access and workspace layout

<a id="sec-1-1"></a>

#### 1.1 Dataset and official sources

[Step 5: primary materials for this study — dataset, corpus, benchmark, or instrument; name/version; **verified** landing URL(s); DOI/portal; license/DUA; citation]

<a id="sec-1-2"></a>

#### 1.2 How to obtain the data

[Step 5: ordered steps from verified instructions only; if not executed this run, subheading **Planned procedure (not yet executed)** — no fake credentials or timings]

<a id="sec-1-3"></a>

#### 1.3 What you receive after download

[Step 5: layout, formats, QC checklist — only after download or quoted from official docs with link; else **TBD until first download** + what goes in `data/<slug>/README.md`]

<a id="sec-1-4"></a>

#### 1.4 Where to put files in this project

[Step 5: paths relative to workspace root (contains `RESEARCH_BRIEF.md`), e.g. `data/<slug>/README.md`, `raw/`, derivatives; no absolute machine paths; no secrets in git]

---

## Study Question and Claim Hierarchy

<a id="sec-2"></a>

### 2. Study question and claim hierarchy

<a id="sec-2-1"></a>

#### 2.1 Primary question

[Step 5: one testable question aligned with rank-1 hypothesis and chosen validation design]

<a id="sec-2-2"></a>

#### 2.2 Confirmatory claims (fixed order)

1. **Primary:** [Main confirmatory claim — what must hold for success]
2. **Key secondary (if any):** [Prespecified secondary; gate order if used]
3. **Robustness:** [What must stay stable under prespecified variations]
4. …

[Add or remove numbered claims to match the study; keep confirmatory order explicit]

<a id="sec-2-3"></a>

#### 2.3 Interpretive bounds

[Step 5: what this design can and cannot conclude — avoid overclaiming beyond evidence type]

---

## Phase 0 — Data and Protocol Lock

<a id="sec-3"></a>

### 3. Phase 0 — Data and protocol lock

[Step 5: feasibility and lock **before** outcome-guided choices]

**Purpose:** [One sentence — e.g. cohort/spec lock, split freeze, preregistration trigger]

1. [Obtain materials under license; archive version, URLs, access date, checksums, codebook]
2. [Availability/QC only — counts, exclusions, missingness; **no** tuning on primary outcomes]
3. [Locked eligibility and analysis population]
4. [Freeze validation splits / holdouts / cross-validation grouping as appropriate to the field]
5. [Preregister or timestamp protocol: contrasts, features, models, multiplicity plan]

**Gate 0:** [Go/no-go criteria; reference §1 for URLs — do not duplicate unverified links]

---

## Phase 1 — Preparation and Analysis-Ready Inputs

<a id="sec-4"></a>

### 4. Phase 1 — Preparation and analysis-ready inputs

[Step 5: domain-specific prep — preprocessing, annotation, feature construction, simulation setup, etc.; all choices locked and documented]

1. [Step]
2. …

**Deliverables:** [QC report, dictionaries, versioned code, immutable ID maps, etc.]

---

## Phase 2 — Primary Analysis Pipeline

<a id="sec-5"></a>

### 5. Phase 2 — Primary analysis pipeline

<a id="sec-5-1"></a>

#### 5.1 Primary endpoint / main test

[Step 5: prespecified main comparison or estimator ladder; same validation folds for all compared conditions]

<a id="sec-5-2"></a>

#### 5.2 Secondary and exploratory analyses

[Step 5: secondary endpoints; exploratory families labeled; complete-case / subset rules]

<a id="sec-5-3"></a>

#### 5.3 Validation design and leakage controls

[Step 5: holdout, nested CV, temporal split, etc.; what is fit only inside training units]

<a id="sec-5-4"></a>

#### 5.4 Metrics, intervals, and multiplicity

[Step 5: primary metrics, calibration or effect sizes as needed; uncertainty method; confirmatory vs exploratory multiplicity]

---

## Phase 3 — Confirmatory Inference

<a id="sec-6"></a>

### 6. Phase 3 — Confirmatory inference

[Step 5: execute locked tests for primary (and gated secondary) claims; report all prespecified confirmatory tests]

1. [Primary claim analysis]
2. [Secondary gate if used]
3. [Prespecified sensitivity tied to confirmatory interpretation]
4. …

---

## Phase 4 — Robustness and Sensitivity

<a id="sec-7"></a>

### 7. Phase 4 — Robustness and sensitivity

[Step 5: run all prespecified robustness checks regardless of primary result; multiverse or sensitivity table]

[Step 5: failure / null-result interpretation — bounded wording when evidence is weak]

---

## Phase 5 — Reporting and Release

<a id="sec-8"></a>

### 8. Phase 5 — Reporting and release

[Step 5: manuscript or report outline mapped to claims; figures list; deviations; sharing per license/DUA]

---

## Required Figures and Tables

<a id="sec-9"></a>

### 9. Required figures and tables

1. [Flow or design schematic]
2. [Primary result figure/table]
3. [Validation schematic]
4. …

[Step 5: one line per required artifact; tie each to a claim or phase]

---

## Preregistered Decision Table

<a id="sec-10"></a>

### 10. Preregistered decision table

| Observed pattern | Permitted conclusion |
|---|---|
| [Pattern A] | [Claim-level wording for primary success] |
| [Pattern B] | [Wording under uncertainty / heterogeneity] |
| [Pattern C] | [Wording for null or inconclusive primary] |
| [Pattern D] | [Wording when sensitivities disagree] |
| … | … |

[Step 5: rows must match §2 claims; no post hoc rows for confirmatory interpretation]

---

## Milestones and Dependencies

<a id="sec-11"></a>

### 11. Milestones and dependencies

| Milestone | Output | Dependency / stop condition |
|---|---|---|
| M0: [Label] | [Artifact] | [Stop or revise if …] |
| M1: [Label] | [Artifact] | [Dependency] |
| M2: [Label] | [Artifact] | [Dependency] |
| … | … | … |

[Step 5: align milestones with Phases 0–5; include prereg and data lock before primary analysis]
