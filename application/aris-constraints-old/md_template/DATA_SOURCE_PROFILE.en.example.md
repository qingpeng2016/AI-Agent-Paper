# Data Source Profile

**Date:** [YYYY-MM-DD]  
**Primary ID:** [manifest `primary_recommendation_id` → block **2.1**]

## Table of contents

1. [Data requirements](#sec-1)
2. [Recommended datasets](#sec-2)
   - [2.1 [d1 — name]](#sec-2-1)
     - [2.1.1 Associated publication](#sec-2-1-1)
     - [2.1.2 Directory structure](#sec-2-1-2)
     - [2.1.3 Access and license](#sec-2-1-3)
     - [2.1.4 Download procedure](#sec-2-1-4)
   - [2.2 [d2 — name] (optional)](#sec-2-2)
     - [2.2.1–2.2.4 …](#sec-2-2-1)
   - [2.3 [d3 — name] (optional)](#sec-2-3)
     - [2.3.1–2.3.4 …](#sec-2-3-1)

---

<a id="sec-1"></a>

## Data requirements

| Dimension | Requirement |
|-----------|-------------|
| Population | [e.g. MDD vs HC] |
| Modality | [e.g. rs-fMRI, T1w] |
| Labels / covariates | [e.g. participants.tsv] |
| Analysis goal | [e.g. graph theory / FC] |
| Minimum N | [number or TBD] |
| Hard exclusions | [e.g. open free direct only] |

---

<a id="sec-2"></a>

## Recommended datasets

<a id="sec-2-1"></a>

### 2.1 [d1 — short name] (primary)

**Repository / stable ID:** …  
**Catalog URL:** …  
**License:** …  
**Sample and acquisition:** N=…; groups …; scanner/tasks …  
**Why primary:** …

<a id="sec-2-1-1"></a>

#### 2.1.1 Associated publication

- Paper: … (DOI …)
- Dataset citation: … `[Dataset]`

<a id="sec-2-1-2"></a>

#### 2.1.2 Directory structure

**Format:** [e.g. BIDS 1.x]

```text
[directory tree — source and date]
```

| Path pattern | Role |
|--------------|------|
| `participants.tsv` | … |
| `sub-*_anat/*_T1w.nii.gz` | … |
| `sub-*_func/*_task-rest_bold.nii.gz` | … |

**structure_verified:** yes — [basis]

<a id="sec-2-1-3"></a>

#### 2.1.3 Access and license

| Check | Status |
|-------|--------|
| License allows reuse | … |
| Free download | … |
| URL verified | … |
| Human subjects / de-id | … |

<a id="sec-2-1-4"></a>

#### 2.1.4 Download procedure

**Method 1 (easiest):** …

**Method 2 (reproducible):** …

**Workspace:** `data-source/raw/<dataset_id>/`  
**Minimal M0 subset:** …

<a id="sec-2-2"></a>

### 2.2 [d2 — short name] (optional)

**Repository / stable ID:** …  
**Sample and acquisition:** …

<a id="sec-2-2-1"></a>

#### 2.2.1 Associated publication

…

<a id="sec-2-2-2"></a>

#### 2.2.2 Directory structure

…

<a id="sec-2-2-3"></a>

#### 2.2.3 Access and license

…

<a id="sec-2-2-4"></a>

#### 2.2.4 Download procedure

…

<a id="sec-2-3"></a>

### 2.3 [d3 — short name] (optional)

**Repository / stable ID:** …  
**Sample and acquisition:** …

<a id="sec-2-3-1"></a>

#### 2.3.1 Associated publication

…

<a id="sec-2-3-2"></a>

#### 2.3.2 Directory structure

…

<a id="sec-2-3-3"></a>

#### 2.3.3 Access and license

…

<a id="sec-2-3-4"></a>

#### 2.3.4 Download procedure

…
