# Paperzilla CLI

A command-line tool for [Paperzilla](https://paperzilla.ai), an AI-powered scientific paper discovery platform. Browse your curated research feed, manage projects, and stay on top of new papers — all from the terminal.

## Install

Build from source (requires Go 1.23+):

```bash
git clone https://github.com/pors/paperzilla-cli.git
cd paperzilla-cli
go build -o pz .
mv pz /usr/local/bin/
```

## Usage

Log in with your Paperzilla account:

```bash
pz login
```

List your projects:

```bash
pz project list
```

Browse your feed:

```bash
pz feed <project-id>
```

```
Machine Learning Papers — 12 papers (total: 142)

★ Must Read  A Novel Approach to Transformer Efficiency
  Smith et al. · arxiv · 2025-08-01 · relevance: 92%

○ Related  On the Convergence Properties of Diffusion Models
  Chen et al. · arxiv · 2025-07-30 · relevance: 74%
```

Filter and export:

```bash
pz feed <project-id> --must-read --limit 5
pz feed <project-id> --since 2025-08-01
pz feed <project-id> --json
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PZ_API_URL` | API base URL | `https://paperzilla.ai` |

## Documentation

Full docs available at [docs/](docs/).

## License

[MIT](LICENSE)
