# Paper skills (ARIS-compatible)

Custom skills ship **inside** `aris-constraints/skills/`. Upstream ARIS skills (e.g. `/research-lit`) still come from the ARIS install; Paper-only skills live here.

| Skill | Command | Contract |
|-------|---------|----------|
| `data-source-discovery` | `/data-source-discovery` | `../data-source-discovery.manifest.example.desc` |

**Project layout after `all-init-project.example.sh`:**

- Canonical: `<project>/aris-constraints/skills/<name>/SKILL.md`
- Mirror: `<project>/skills/<name>/SKILL.md` (so ARIS can resolve `/data-source-discovery`)

Manual install into an existing project:

```bash
cp -R /path/to/AI-Agent-Paper/application/aris-constraints/skills/* ~/my-paper/aris-constraints/skills/
mkdir -p ~/my-paper/skills
cp -R ~/my-paper/aris-constraints/skills/* ~/my-paper/skills/
```

Run:

```bash
cd ~/my-paper
bash aris-constraints/data-source-discovery.aris.example.sh
bash aris-constraints/data-source-discovery.validate.example.sh
```
