# Contributing

Argus uses GitHub to manage reviews of pull requests.

- If you are a new contributor see: [Steps to Contribute](#steps-to-contribute)

- If you have a trivial fix or improvement, go ahead and create a pull request.

- If you plan to do something more involved, investigate open [issues](https://github.com/release-argus/Argus/issues) to see whether others are planning to work on this issue and open one if your search is empty.

- Relevant coding style guidelines are the [Go Code Review Comments](https://code.google.com/p/go-wiki/wiki/CodeReviewComments) and the _Formatting and style_ section of Peter Bourgon's [Go: Best Practices for Production Environments](https://peter.bourgon.org/go-in-production/#formatting-and-style).

## Steps to Contribute

Should you wish to work on an issue, please claim it first by commenting on the GitHub issue that you want to work on it. This is to prevent duplicated efforts from contributors on the same issue.

For complete instructions on how to compile see: [Building From Source](https://github.com/release-argus/Argus#building-from-source)

For quickly compiling and testing your changes do:

```
# For building.
make go-build
./argus
```

```
# For testing.
go test ./...
```

## Pre-commit Hooks

If you're not making a small change, please set up and run the pre-commit checkers.

To check the commits, we use [Pre-Commit-GoLang](https://github.com/tekwizely/pre-commit-golang) and [Husky](https://typicode.github.io/husky/#/).

Pre-Commit-Golang:

- Simply run `bash .pre-commit-config.requirements.sh` and the various GoLang modules required will be installed.
- After this, run `pre-commit run --all-files`

Husky:

- Simply `npm install` in the root directory.
  (This checks that the commit message matches the [Conventional Commits](https://www.conventionalcommits.org) standard).

## Pull Request Checklist

- Branch from the master branch and, if needed, rebase to the current master branch before submitting your pull request. If it doesn't merge cleanly with master you may be asked to rebase your changes.

- Commits should be as small as possible, while ensuring that each commit is correct independently (i.e., each commit should compile and pass tests).

- Add tests relevant to the fixed bug or new feature.
