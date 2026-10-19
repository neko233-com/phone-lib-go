# Contributing

Thanks for taking the time to contribute to `phone-lib-go`.

## Development

Requirements: Go 1.25 or newer.

```bash
go test ./...        # unit tests
go vet ./...         # static analysis
go test -race ./...  # race detector (requires a C compiler)
gofmt -w .           # formatting
```

The race detector needs a C compiler on Windows; it runs on Linux in CI.

## Pull requests

- Keep changes focused and add tests for new behavior.
- Run `gofmt -w .`, `go vet ./...` and `go test ./...` before opening a PR.
- Update `README.md` / `CHANGELOG.md` when user-facing behavior changes.
- Do not edit embedded data files by hand. `internal/cndata/phone.dat` is an
  upstream snapshot (see `internal/cndata/NOTICE.md`); describe the data update
  and its version in the PR instead.

## Release process

Releases are tag-driven:

1. Bump `phone.Version` in `version.go` and move the `CHANGELOG.md` section from
   `Unreleased` to the new version.
2. Commit and push to `main`.
3. Create an **annotated** tag whose message is the release note:
   `git tag -a vX.Y.Z -m "..."` and `git push origin vX.Y.Z`.
   If the note contains Markdown headings (lines starting with `#`), add
   `--cleanup=verbatim` and use `-F notes.md`; otherwise Git strips those lines
   as comments.
4. `.github/workflows/release.yml` verifies that the tag matches
   `phone.Version`, runs the race tests, and creates the GitHub release from the
   tag message. To publish or re-run an existing tag, use **Actions → Release →
   Run workflow** and enter the tag (for example `v0.2.1`).

## Code of conduct

Be respectful and constructive. Harassment or personal attacks are not
tolerated.

