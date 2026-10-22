# Public data repositories (discovery reference)

Use during **Phase 2–3** of `/data-source-discovery`. Prefer repositories with **clear license**, **DOI or stable ID**, and **documented free download** (no paid wall, no consortium-only gate for the cohort you need).

| Domain | Repository | ID pattern | Typical license | Free access notes |
|--------|------------|------------|-----------------|-------------------|
| Neuroimaging (BIDS) | [OpenNeuro](https://openneuro.org) | `ds######` | CC0 (after grace) | S3/GitHub; accept terms on download |
| Neuroimaging | [OpenNeuro PET](https://openneuro.org) | `ds######` | varies | same portal |
| Multi-modal brain | [NITRC](https://www.nitrc.org/) | project pages | varies | some open, some request |
| UK Biobank | [UK Biobank](https://www.ukbiobank.ac.uk/) | application | custom DUA | **not** free anonymous — exclude from `primary_recommendation` unless user already has access |
| ADNI | [ADNI](https://adni.loni.usc.edu/) | application | custom | registration required — secondary only |
| General research data | [Zenodo](https://zenodo.org) | DOI `10.5281/zenodo.*` | often CC-BY / CC0 | direct HTTP |
| General | [Figshare](https://figshare.com) | DOI | often CC-BY | direct |
| NIH | [NDA](https://nda.nih.gov/) | GUID | controlled | **not** free anonymous for most — secondary |
| fMRI prep atlases | GitHub releases | repo tag | varies | atlases are not subject BOLD |
| ML vision/NLP | [Hugging Face Datasets](https://huggingface.co/datasets) | `org/name` | per dataset card | check card `license` |
| ML | [UCI ML Repository](https://archive.ics.uci.edu/) | name | varies | usually free HTTP |
| ML | [OpenML](https://www.openml.org/) | dataset id | varies | API + download |
| Earth / geo | [NASA Earthdata](https://www.earthdata.nasa.gov/) | product | varies | often free account |
| Social / survey | [ICPSR](https://www.icpsr.umich.edu/) | study number | varies | often institutional |
| COVID / health | [PhysioNet](https://physionet.org/) | project | varies | some open, some credentialed |
| Cancer imaging | [TCIA](https://www.cancerimagingarchive.net/) | collection | often CC | NBIA downloader |

**License allow-list for `access_tier: open_free_direct` success:** CC0, CC-BY (2.0/3.0/4.0), CC-BY-SA, Open Data Commons (PDDL, ODbL with attribution), US Public Domain, explicit "public domain" on official page.

**Exclude from primary recommendation:** "application required", "DUA", "embargo", "commercial license fee", "members only", unknown license.

**Human subjects:** note de-identification, terms of reuse (e.g. OpenNeuro download agreement), and cite original acquisition paper + dataset DOI in the profile.
