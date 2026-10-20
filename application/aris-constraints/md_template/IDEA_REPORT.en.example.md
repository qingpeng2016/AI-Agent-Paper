# Idea Discovery Report: {short topic title}

<!--
  Target: idea-stage/IDEA_REPORT.md (composed across Steps 1–5).
  Pair: IDEA_REPORT.zh.example.md (same sec-* anchors for TOC jumps).
  Top-level report: 6 chapters (sec-1 … sec-6). Subsections use 1 / 1.1 / 1.1.1 under each contract H2.
  Contract H2 lines must stay exactly as written (manifest + validate + platform gate).
  Do not hard-code datasets, URLs, or sample sizes. Remove this comment when done.
-->

## Table of contents

- [1. Executive summary](#sec-1)
- [2. Literature landscape](#sec-2)
  - [2.1 Scope and heterogeneity](#sec-2-1)
  - [2.2 Key findings](#sec-2-2)
  - [2.3 Gaps identified](#sec-2-3)
  - [2.4 Paper entries](#sec-2-4)
- [3. Recommended ideas](#sec-3)
  - [3.1 Idea 1](#sec-3-1)
  - [3.2 Idea 2](#sec-3-2)
  - [3.3 Idea 3](#sec-3-3)
- [4. Novelty check](#sec-4)
  - [4.1 Overview](#sec-4-1)
  - [4.2 Per-idea checks](#sec-4-2)
  - [4.3 Portfolio synthesis](#sec-4-3)
- [5. External review](#sec-5)
  - [5.1 Overall verdict](#sec-5-1)
  - [5.2 Strengths](#sec-5-2)
  - [5.3 Critical weaknesses](#sec-5-3)
  - [5.4 Method concerns](#sec-5-4)
  - [5.5 Venue fit and objections](#sec-5-5)
  - [5.6 Actionable follow-ups and synthesis](#sec-5-6)
- [6. References](#sec-6)

---

## Executive Summary

<a id="sec-1"></a>

### 1. Executive summary

{Step 5: rank-1 refinement, next milestones, data/prereg status; 2–4 paragraphs.}

#### 1.1 Report metadata

**Research direction:** {from RESEARCH_BRIEF.md}

**Target venue:** {venue}

**Generated:** {YYYY-MM-DD}

**Ideas evaluated:** {generated} → {after filters} → {pilots run} → {recommended}

---

## Literature Landscape

<a id="sec-2"></a>

### 2. Literature landscape

<a id="sec-2-1"></a>

#### 2.1 Scope and heterogeneity

{Step 1: consensus, sources of variation; no unverified URLs.}

<a id="sec-2-2"></a>

#### 2.2 Key findings ({N} papers, {year range})

1. {finding + citable basis (author/year); or TBD}
2. …

<a id="sec-2-3"></a>

#### 2.3 Gaps identified

- {gap 1}
- …

<a id="sec-2-4"></a>

#### 2.4 Paper entries

{Each manifest papers[] entry: at least one urls[] string must appear verbatim in this file.}

<a id="sec-2-4-1"></a>

##### 2.4.1 {p1 short title} ({year})

- **Takeaway:** {1–2 sentences}
- **Link:** {url or doi from manifest}

{Repeat 2.4.x for each paper: 2.4.2, 2.4.3, …}

---

## Recommended Ideas (ranked)

<a id="sec-3"></a>

### 3. Recommended ideas

{Optional 1–2 sentence portfolio intro.}

<a id="sec-3-1"></a>

### Idea 1: {byte-identical to idea-creator.manifest ideas[].title where rank=1}

##### 3.1.1 Methods and hypothesis

**Method (what we would actually do):**
1. {step}
2. …

**Hypothesis:** {…}

##### 3.1.2 Experiment and expected outcomes

**Minimum experiment:** {…}

**Expected outcomes:**
- Positive signal: {…}
- Null/negative: {…}

##### 3.1.3 Assessment and risks

**Novelty:** {score}/10 — {…}

**Feasibility:** {…}

**Risks:** {…}

**Contribution type:** {…}

**Pilot results:** {…}

**Likely reviewer objection:** {…}

**Why we should do this:** {…}

---

<a id="sec-3-2"></a>

### Idea 2: {title from manifest rank=2}

{Same 3.1.1–3.1.3 structure with numbering 3.2.1–3.2.3, or reuse 3.1.x labels under each Idea block.}

---

<a id="sec-3-3"></a>

### Idea 3: {title from manifest rank=3}

{Same structure.}

---

## Novelty Verification

<a id="sec-4"></a>

### 4. Novelty check

<a id="sec-4-1"></a>

#### 4.1 Overview

{Step 3: portfolio verdict, execution priority.}

<a id="sec-4-2"></a>

#### 4.2 Per-idea checks

##### 4.2.1 Idea 1: {title}

**Novelty score:** {n}/10  
**Recommendation:** {PROCEED / CAUTIOUS / ABANDON, etc.}

**Closest prior work:**
- {citation} — {gap vs this idea}

**Main risks:** {…}

##### 4.2.2 Idea 2: {title}

{Same.}

##### 4.2.3 Idea 3: {title}

{Same.}

<a id="sec-4-3"></a>

#### 4.3 Portfolio synthesis

{…}

---

## External Critical Review

<a id="sec-5"></a>

### 5. External review

{Step 4: rank-1 idea; idea_title matches manifest.}

<a id="sec-5-1"></a>

#### 5.1 Overall verdict

{…}

<a id="sec-5-2"></a>

#### 5.2 Strengths

- {…}

<a id="sec-5-3"></a>

#### 5.3 Critical weaknesses

1. {…}

<a id="sec-5-4"></a>

#### 5.4 Method concerns

- {…}

<a id="sec-5-5"></a>

#### 5.5 Venue fit and reviewer objections

{…}

<a id="sec-5-6"></a>

#### 5.6 Actionable follow-ups and reviewer synthesis

{…}

---

<a id="sec-6"></a>

### 6. References

{Each entry: authors. **Title**. *Journal*. year;volume(issue):pages. https://doi.org/… — URLs must match §2.4 and appear verbatim from manifest `papers[].urls`.}

Example:

1. Yang Q, Fu L, Xia M, et al. Disrupted intrinsic functional brain topology in patients with major depressive disorder. *Mol Psychiatry*. 2022;27(1):854–866. https://doi.org/10.1038/s41380-021-01247-2
