# Data Source Profile

**Brief anchor:** [one line from RESEARCH_BRIEF.md]  
**Date:** [YYYY-MM-DD]  
**Primary dataset ID:** [manifest datasets[].id]

---

## Research data needs

| Dimension | Requirement |
|-----------|-------------|
| Population | [e.g. MDD vs healthy controls] |
| Modality | [e.g. rs-fMRI, T1w] |
| Labels / covariates | [e.g. diagnosis in participants.tsv] |
| Analysis goal | [e.g. graph theory on resting-state FC] |
| Minimum N (estimate) | [number or TBD with rationale] |

---

## Candidate datasets

| ID | Name | Repository | access_tier | Fit | License |
|----|------|------------|-------------|-----|---------|
| d1 | … | OpenNeuro | open_free_direct | high | CC0-1.0 |

---

## Primary recommendation

One block per dataset: **`### 1.`**, **`### 2.`**, … (`primary_recommendation_id` → block **1**). The **`####`** subsections below belong **under that dataset**, not as document-level H2.

### 1. [d1 — short name]

**Name:** …  
**Stable ID / DOI:** …  
**Catalog URL:** …  
**Why primary:** …

#### Cohort and design

- **N:** … (groups: …)
- **Scanner / protocol:** … (if known)
- **Tasks:** …

#### Associated publication

- Paper: … (DOI …)
- Dataset citation: … `[Dataset]`

#### Directory and file structure

**Format:** [e.g. BIDS 1.x]

```text
[directory tree — verified source: README URL or listing date]
```

| Path pattern | Role |
|--------------|------|
| `participants.tsv` | … |
| `sub-*_anat/*_T1w.nii.gz` | … |
| `sub-*_func/*_task-rest_bold.nii.gz` | … |

**structure_verified:** yes — [basis: doc URL / probe path / date]

#### Access and license

| Check | Status |
|-------|--------|
| License allows research reuse | … |
| Free download (no purchase) | … |
| URL verified | … (method, date) |
| Human subjects / de-id | … |
| User IRB for secondary use | confirm with institution |

#### Download procedure

##### Method 1: … (easiest)
1. …

##### Method 2: … (recommended for reproducibility)
1. …

**Suggested workspace layout:** `data-source/raw/<dataset_id>/` (BIDS root)

**Minimal M0 subset:** …

### 2. [d2 — …] (optional; same **`####`** structure when filled)

**Name:** …  
**Stable ID / DOI:** …  
**Catalog URL:** …  
**Why listed here:** …

#### Cohort and design

…

#### Associated publication

…

#### Directory and file structure

…

#### Access and license

…

#### Download procedure

…
