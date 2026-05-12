# Paperzilla CLI

A command-line tool for [Paperzilla](https://paperzilla.ai), an AI-powered scientific paper discovery platform. Read canonical papers, browse project recommendations, manage projects, and stay on top of new papers from the terminal.

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
pz project list --json
```

`pz project list --json` returns a compact summary array with the same fields shown in the table output: `id`, `name`, `mode`, and `visibility`.

Show a single project:

```bash
pz project <project-id>
pz project <project-id> --json
```

`pz project <project-id> --json` returns the full project record, including `positive_keywords`, `negative_keywords`, watched `sources`, and watched `categories`.

Read a canonical paper by Paperzilla paper ID:

```bash
pz paper <paper-id>
pz paper <paper-id> --json
pz paper <paper-id> --markdown
```

Show that paper in the context of one of your projects:

```bash
pz paper <paper-id> --project <project-id>
```

Open a recommendation from your feed:

```bash
pz rec <project-paper-id>
pz rec <project-paper-id> --json
pz rec <project-paper-id> --markdown
```

Leave recommendation feedback:

```bash
pz feedback <project-paper-id> upvote
pz feedback <project-paper-id> upvote --json
pz feedback <project-paper-id> star
pz feedback <project-paper-id> downvote --reason not_relevant
pz feedback clear <project-paper-id>
pz feedback clear <project-paper-id> --json
```

`clear` is a subcommand, so the valid syntax is `pz feedback clear <project-paper-id>`, not `pz feedback <project-paper-id> clear`.
`pz feedback --json` returns the feedback object. `pz feedback clear --json` returns a small confirmation envelope because the backend clear endpoint returns `204 No Content`.

Canonical `pz paper --markdown` only returns markdown when it is already prepared. `pz rec --markdown` can queue markdown generation and prints a friendly message if it is still being prepared.

Get a project ID, then browse or search its feed:

```bash
pz project list
pz feed <project-id>
pz feed search --project-id <project-id> --query "latent retrieval"
```

```
Machine Learning Papers — 12 papers (total: 142)

★ Must Read [★]  A Novel Approach to Transformer Efficiency
  Smith et al. · arxiv · 2025-08-01 · relevance: 92%

○ Related [↓]  On the Convergence Properties of Diffusion Models
  Chen et al. · arxiv · 2025-07-30 · relevance: 74%
```

Recommendations can show existing feedback inline:

- `[↑]` upvote
- `[↓]` downvote
- `[★]` star

Filter and export:

```bash
pz feed <project-id> --must-read --limit 5
pz feed <project-id> --since 2025-08-01
pz feed <project-id> --json
```

Search the full feed:

```bash
pz feed search --project-id <project-id> --query "latent retrieval"
pz feed search --project-id <project-id> --query "Proxi" --feedback-filter starred
pz feed search --project-id <project-id> --query "latent retrieval" --must-read --limit 10 --offset 20
pz feed search --project-id <project-id> --query "latent retrieval" --json
```

Use `pz project list` first if you need to look up the project ID.
Text output includes the echoed query, returned item count, and whether more ranked results are available via `has_more`.
Search uses server-side ranking across the full feed, not just already loaded browse pages.
Supported `--feedback-filter` values are `all`, `unrated`, `liked`, `disliked`, `starred`, `not-relevant`, and `low-quality`.
Queries are trimmed and must be 3-200 characters.

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
