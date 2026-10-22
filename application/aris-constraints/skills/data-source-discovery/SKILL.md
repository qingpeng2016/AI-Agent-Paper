---
name: data-source-discovery
description: 'Find public research datasets that match a study brief: verify free legal download, document license and ethics reuse, explain directory/schema and what each modality is for. Use before /research-lit or /experiment-plan when the paper depends on open data, or when the user asks for dataset search, OpenNeuro, BIDS, or data feasibility.'
argument-hint: "[optional focus — defaults to RESEARCH_BRIEF.md]"
allowed-tools: Bash(*), Read, Write, Edit, Grep, Glob, WebSearch, WebFetch
---

# Data Source Discovery: Feasible, Open, Structured

Discover and document datasets for: **$ARGUMENTS**

## Overview

This skill runs **before** idea-heavy planning when empirical work needs **real, downloadable data**. Output is a **data anchor**: what exists, where to get it for free, what files mean, and what is still missing for the stated research goals.

**Success means:** at least one **primary** dataset that is **public**, **license-clear for reuse**, **free to download without purchase**, with **verified landing/download URLs** and **schema/structure checked against official docs or a listing** — not guessed from memory.

If the Paper workspace defines `aris-constraints/data-source-discovery.manifest.example.desc`, **follow that contract for paths, JSON, and H2 sections** in addition to this skill.

## Constants

- **INPUT_BRIEF = `RESEARCH_BRIEF.md`** — Required unless `$ARGUMENTS` fully specifies domain, modality, cohort, and task.
- **OUTPUT_DIR = `data-source/`** — **All** skill outputs (manifest, profiles). Run **`mkdir -p data-source`** first if missing.
- **DATA_STAGING = `data-source/raw/`** — Optional probe downloads (small files only unless user approves size).
- **MAX_CANDIDATES = 8** — Internal search shortlist breadth (not all appear in profile).
- **MIN_PROFILE_DATASETS = 1** · **MAX_PROFILE_DATASETS = 3** — **`## 推荐数据集`** documents **`### 2.1`–`2.3`** only (at least **2.1**; up to three ranked entries).
- **MAX_PRIMARY = 1** — One locked **`primary_recommendation_id`** (always **`### 2.1`** / `datasets[0]` or matching id).
- **SOURCES** — Parse `— sources:` from the invocation. Default: `web`. Recommended for discovery: `web, gemini, openalex` when available.
- **PROBE_DOWNLOAD = false** — When `true`, download a **minimal** subset (metadata + 1–2 subjects or &lt;50MB) to validate layout.
- **ALLOW_GATED_SECONDARY = true** — List apply-to-access datasets only under **Secondary / not free**, never as primary.

> Overrides:
> - `/data-source-discovery — sources: web, openalex`
> - `/data-source-discovery — probe download: true`
> - `/data-source-discovery "rs-fMRI MDD graph theory" — sources: web, gemini`

Repository hints: read `aris-constraints/skills/data-source-discovery/repositories-reference.md` (project root; after init lives beside other `aris-constraints/` contracts).

## Workflow

### Phase 0: Parse research needs

Read `RESEARCH_BRIEF.md` (and `$ARGUMENTS`). Extract a **Data Requirements Table** (internal, then reflect in outputs):

| Dimension | What to fix |
|-----------|-------------|
| Population / units | e.g. MDD vs HC, species, cell line |
| Modality | e.g. rs-fMRI, T1, behavior, text |
| Labels / covariates | diagnosis, HAMD, age, sex |
| Task | rest, task name, stimulus |
| Analysis intent | graph theory, classification, survival |
| Minimum N | rough floor for power (honest if unknown) |
| Hard exclusions | no application-only, no paid-only, no GDPR-restricted if user lacks access |

If requirements are ambiguous, state **assumptions** explicitly in the profile — do not silently narrow the brief.

### Phase 1: Search and shortlist

Search prioritized repositories for the domain (see repositories reference). For each candidate record:

- Official catalog page + **DOI or stable ID** (OpenNeuro `ds######`, Zenodo DOI, HF dataset id, etc.)
- **Associated publication** (first paper describing acquisition)
- **Sample size** and group breakdown if published
- **License** text from the official page (not blogs)

Drop candidates that clearly violate **free + public** unless listing as secondary with reason.

Use **WebFetch / WebSearch**; prefer official provider pages over third-party mirrors unless the mirror is the documented distribution (e.g. OpenNeuro S3).

### Phase 2: Access and license gate (mandatory per candidate)

For each shortlisted dataset classify **`access_tier`**:

| Tier | Meaning | Primary? |
|------|---------|----------|
| `open_free_direct` | Anonymous or click-through terms only; no fee; HTTP/S3/git clone documented | Yes |
| `free_registration` | Free account or application, no purchase | Secondary only |
| `restricted_dua` | Consortium / IRB / DUA | Secondary only |
| `paid_or_unknown` | Fee or unclear license | Exclude or flag reject |

Record **`license_spdx_or_name`** (e.g. `CC0-1.0`, `CC-BY-4.0`) and **`reuse_notes`**: attribution, commercial use, share-alike, human-subject redistribution terms.

**Verify URLs** in this run when network allows: `curl -I`, `WebFetch`, or small test GET. Record **verification_status** and date. Do not claim “verified downloadable” without evidence.

### Phase 3: Structure and semantics (what files are for)

For each candidate (especially primary), document:

1. **Format** — BIDS, CSV, NIfTI, parquet, etc.
2. **Directory tree** — top 2–3 levels from official README/datacard or **real listing** (`aws s3 ls`, GitHub tree, downloaded `dataset_description.json`).
3. **Key files table** — path pattern → role (e.g. `sub-*_task-rest_bold.nii.gz` = rs-fMRI time series; `*_T1w.nii.gz` = structural for registration).
4. **Metadata tables** — `participants.tsv` columns, label files, JSON sidecars.
5. **Gaps vs Phase 0** — missing modality, missing clinical scale, insufficient N.
6. **Minimal analysis-ready subset** — what to download first for M0/sanity (e.g. `participants.tsv` + 2 subjects anat+func).

**Forbidden:** inventing column names or paths not in docs or a listing you inspected.

If **PROBE_DOWNLOAD = true**, place files under `data-source/raw/<dataset_id>/`, reconcile profile text with `ls` / sample read, note **disk_paths** in manifest.

### Phase 4: Rank and lock primary

Score fit (high/medium/low) against Phase 0. Choose **`primary_recommendation`**:

- Must be `access_tier: open_free_direct`
- License allow-list satisfied
- Structure section complete with **structure_verified: true**
- Gaps must be **survivable** for a plausible paper OR explicitly warn that `/research-lit` should narrow claims

Write **1–3 datasets** into the profile (`### 2.1` required; `2.2` / `2.3` if useful alternates). **`datasets[]` in manifest** must match profile count (1–3). Only **`2.1`** may be `open_free_direct` primary; `2.2`/`2.3` may be lower tier with reason in lead.

### Phase 5: Write deliverables

#### 5.1 `data-source/DATA_SOURCE_PROFILE.en.md`

**Mandatory:** Read `aris-constraints/md_template/DATA_SOURCE_PROFILE.en.example.md` and **clone its heading tree** (TOC + `sec-*` anchors). Replace placeholders only.

- **`## Data requirements`** → **`## Recommended datasets`** (only two document H2).
- **1–3 blocks:** **`### 2.1`** (primary) through optional **`### 2.3`**; each block has **`#### 2.x.1`–`2.x.4`** (publication, directory structure, access/license, download). Put sample size / protocol in each **`### 2.x` lead**.
- **Anti-patterns:** extra H2 (candidates, “detailed version”, gaps, downstream); duplicating sections after the template; `### 1.` legacy blocks.

#### 5.2 `data-source/DATA_SOURCE_PROFILE.zh.md`

Same tree as EN: **`## 数据需求`**, **`## 推荐数据集`**, **`### 2.1`**, **`#### 2.1.1 关联发表`** … **`2.1.4 下载步骤`**. Simplified Chinese **专业术语**; IDs/paths aligned with EN.

#### 5.2b Validate before success

From project root: `bash aris-constraints/data-source-discovery.validate.example.sh`. Do not set manifest `result: success` until profiles pass.

#### 5.3 `data-source/data-source-discovery.manifest.result.json`

Write JSON per Paper manifest contract (`datasets[]`, `primary_recommendation_id`, `executive_summary`, `result`).

#### 5.4 Optional IDEA_REPORT §1 anchor

If `idea-stage/IDEA_REPORT.en.md` has `## Executive Summary` / data subsection, add a short **Data anchor** paragraph pointing to the profile and primary ID — only when Paper manifest asks for it.

#### 5.5 User summary (chat)

```
Data source discovery complete.

Primary (free & open): [name] ([id])
License: [license] | Verified: [date/method]
Minimal download: [bullets]
Gaps: [top 1–2]

Output dir: data-source/
Profiles: data-source/DATA_SOURCE_PROFILE.en.md, .zh.md
Manifest: data-source/data-source-discovery.manifest.result.json
```

## Key rules

- **No primary without free direct access** — registration-only datasets are never primary for this skill’s success gate.
- **Verify, don’t guess** — URLs and schemas need a recorded check method.
- **Separate atlas/resources from subject data** — e.g. Schaefer atlas ≠ OpenNeuro BOLD runs; link atlas as **external resource**.
- **Cite for publication** — include recommended **dataset citation** and **original study paper** in the profile.
- **Ethics** — human data: note de-identification, redistribution terms, no re-identification; IRB is the user’s institution — state “confirm locally” when unsure.
- **Do not download huge cohorts** unless user requested or `PROBE_DOWNLOAD` with stated size cap.
- **Large writes** — if Write fails, use Bash heredoc chunks.

## Composing with other skills

```
/data-source-discovery  →  data anchor (this skill)
/research-lit           →  gap-focused literature using locked dataset
/idea-creator           →  ideas constrained to available fields/modalities
/experiment-plan        →  acquisition section copies verified paths from profile
/experiment-bridge      →  after data on disk per plan
```

Recommended Paper order: **`RESEARCH_BRIEF` → `/data-source-discovery` → `/research-lit` → …**
