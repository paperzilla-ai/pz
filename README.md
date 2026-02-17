# Paperzilla CLI

A command-line tool for [Paperzilla](https://paperzilla.ai), an AI-powered scientific paper discovery platform. Browse your curated research feed, manage projects, and stay on top of new papers — all from the terminal.

## Install

### macOS

```bash
brew install paperzilla-ai/tap/pz
```

### Windows

Via [Scoop](https://scoop.sh):

```bash
scoop bucket add paperzilla-ai https://github.com/paperzilla-ai/scoop-bucket
scoop install pz
```

### Linux

Download the latest binary from [GitHub Releases](https://github.com/paperzilla-ai/pz/releases):

```bash
curl -sL https://github.com/paperzilla-ai/pz/releases/latest/download/pz_linux_amd64.tar.gz | tar xz
sudo mv pz /usr/local/bin/
```

### Build from source

Requires Go 1.23+:

```bash
git clone https://github.com/paperzilla-ai/pz.git
cd pz
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

### Subscribe in a feed reader

Get an Atom feed URL you can add to any feed reader ([Vienna RSS](https://github.com/ViennaRSS/vienna-rss), NetNewsWire, Feedly, etc.):

```bash
pz feed <project-id> --atom
```

This prints a URL with an embedded feed token. Paste it into your feed reader to subscribe — no login required on the reader side. The token is per-user and the same URL is returned on repeated calls. Running `--atom` again after revoking will generate a new token.

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PZ_API_URL` | API base URL | `https://paperzilla.ai` |

## Documentation

Full docs available at [docs/](docs/).

## License

[MIT](LICENSE)
