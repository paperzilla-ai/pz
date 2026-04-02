# Paperzilla CLI

A command-line tool for [Paperzilla](https://paperzilla.ai), an AI-powered scientific paper discovery platform. Browse your curated research feed, manage projects, and stay on top of new papers — all from the terminal.

New to `pz`? Start here: [docs.paperzilla.ai/guides/cli-getting-started](https://docs.paperzilla.ai/guides/cli-getting-started)

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

## Update

Run `pz update` to check whether your CLI is on the latest release and see upgrade instructions based on how it was installed.

```bash
pz update
```

`pz update` auto-detects common install methods:

- Homebrew
- Scoop
- GitHub release binaries
- Source builds

If detection is ambiguous, override it explicitly:

```bash
pz update --install-method homebrew
pz update --install-method scoop
pz update --install-method release
pz update --install-method source
```

Supported values are `auto`, `homebrew`, `scoop`, `release`, and `source`.

On interactive terminals, `pz` also prints a colored footer after successful commands when you are behind the latest release, with a prompt to run `pz update`.

### macOS

```bash
brew update
brew upgrade pz
```

### Windows

```bash
scoop update pz
```

### Linux

If you installed from GitHub Releases, download the latest binary again:

```bash
curl -sL https://github.com/paperzilla-ai/pz/releases/latest/download/pz_linux_amd64.tar.gz | tar xz
sudo mv pz /usr/local/bin/
```

If you installed from source, pull the latest code and rebuild:

```bash
git pull
go build -o pz .
sudo mv pz /usr/local/bin/
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

Show a single paper or feed item:

```bash
pz paper <paper-or-feed-id>
pz paper <paper-or-feed-id> --json
pz paper <paper-or-feed-id> --markdown
```

`--markdown` prints raw paper markdown to stdout. If the markdown is still being prepared, `pz` prints a friendly message and asks you to try again in a minute or so.

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

Full docs available at [docs.paperzilla.ai](https://docs.paperzilla.ai/guides/cli-getting-started).

## License

[MIT](LICENSE)
