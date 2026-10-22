# Paper skills (ARIS-compatible)

Custom skills ship **inside** `aris-constraints/skills/<name>/` as **bundles** (contract + launcher + validate; Paper-only skills also include `SKILL.md`). Upstream ARIS skills (e.g. `/research-lit`) come from the ARIS install; Paper **constraints** live next to each skill name below.

| Skill | Command | Bundle directory |
|-------|---------|------------------|
| `data-source-discovery` | `/data-source-discovery` | `data-source-discovery/` (+ `SKILL.md`) |
| `research-lit` | `/research-lit` | `research-lit/` (Step 1) |
| `idea-creator` | `/idea-creator` | `idea-creator/` (Step 2) |
| `novelty-check` | `/novelty-check` | `novelty-check/` (Step 3) |
| `research-review` | `/research-review` | `research-review/` (Step 4) |
| `research-refine-pipeline` | `/research-refine-pipeline` | `research-refine-pipeline/` (+ `*.explain.desc`, Step 5) |
| `research-refine` | `/research-refine` | `research-refine/` (standalone refine) |
| `experiment-plan` | `/experiment-plan` | `experiment-plan/` (standalone plan) |

Each bundle typically contains:

- `<name>.desc` — paths, JSON, report gates  
- `<name>.aris.sh` — `aris prompt '/<name> …'`  
- `<name>.validate.sh` — deterministic CI check  

**Project layout after `all-init-project.example.sh`:**

- Canonical: `<project>/aris-constraints/skills/<name>/`
- Mirror: `<project>/skills/<name>/` (copy **entire** directory for ARIS slash commands)

Manual install:

```bash
cp -R /path/to/AI-Agent-Paper/application/aris-constraints/skills/* ~/my-paper/aris-constraints/skills/
mkdir -p ~/my-paper/skills
cp -R ~/my-paper/aris-constraints/skills/* ~/my-paper/skills/
```

**Idea discovery pipeline (Steps 1–5):**

```bash
cd ~/my-paper
bash aris-constraints/skills/research-lit/research-lit.aris.sh
bash aris-constraints/skills/idea-creator/idea-creator.aris.sh
bash aris-constraints/skills/novelty-check/novelty-check.aris.sh
bash aris-constraints/skills/research-review/research-review.aris.sh
bash aris-constraints/skills/research-refine-pipeline/research-refine-pipeline.aris.sh
```

Validate after each step: `bash aris-constraints/skills/<name>/<name>.validate.sh`

**Standalone:**

```bash
bash aris-constraints/skills/research-refine/research-refine.aris.sh
bash aris-constraints/skills/experiment-plan/experiment-plan.aris.sh
bash aris-constraints/skills/data-source-discovery/data-source-discovery.aris.sh
```
