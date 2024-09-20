# go-sequencing

go-sequencing defines a generic sequencing interface for modular blockchains

<!-- markdownlint-disable MD013 -->
[![build-and-test](https://github.com/rollkit/go-sequencing/actions/workflows/ci_release.yml/badge.svg)](https://github.com/rollkit/go-sequencing/actions/workflows/ci_release.yml)
[![golangci-lint](https://github.com/rollkit/go-sequencing/actions/workflows/lint.yml/badge.svg)](https://github.com/rollkit/go-sequencing/actions/workflows/lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rollkit/go-sequencing)](https://goreportcard.com/report/github.com/rollkit/go-sequencing)
[![codecov](https://codecov.io/gh/rollkit/go-sequencing/branch/main/graph/badge.svg?token=CWGA4RLDS9)](https://codecov.io/gh/rollkit/go-sequencing)
[![GoDoc](https://godoc.org/github.com/rollkit/go-sequencing?status.svg)](https://godoc.org/github.com/rollkit/go-sequencing)
<!-- markdownlint-enable MD013 -->

## Sequencing Interface

<!-- markdownlint-disable MD013 -->
| Method        | Params                                                   | Return          |
| ------------- | -------------------------------------------------------- | --------------- |
| `SubmitRollupTransaction` | `context.Context`, `rollupId RollupId`, `tx Tx` | `error` |
| `GetNextBatch` | `context.Context`, `lastBatchHash Hash` | `batch Batch`, `timestamp Time` `error` |
| `VerifyBatch` | `context.Context`, `batchHash Hash` | `success bool`, `error` |
<!-- markdownlint-enable MD013 -->

Note: `Batch` is []`Tx` and `Tx` is `[]byte`. Also `Hash` and `RollupId` are `[]byte`.

## Implementations

The following implementations are available:

* [centralized-sequencer][centralized] implements a centralized sequencer that
  posts rollup transactions to Celestia DA.
* [astria-sequencer][astria] implements a Astria sequencer middleware that
  connects to Astria shared sequencer.

In addition the following helper implementations are available:

* [DummySequencer][dummy] implements a Mock sequencer useful for testing.
* [Proxy][proxy] implements a proxy server that forwards requests to a gRPC
  server. The proxy client can be used directly to interact with the sequencer
  service.

## Helpful commands

```sh
# Generate protobuf files. Requires docker.
make proto-gen

# Lint protobuf files. Requires docker.
make proto-lint

# Run tests.
make test

# Run linters (requires golangci-lint, markdownlint, hadolint, and yamllint)
make lint
```

## Contributing

We welcome your contributions! Everyone is welcome to contribute, whether it's
in the form of code, documentation, bug reports, feature
requests, or anything else.

If you're looking for issues to work on, try looking at the [good first issue
list][gfi].  Issues with this tag are suitable for a new external contributor
and is a great way to find something you can help with!

Please join our
[Community Discord](https://discord.com/invite/YsnTPcSfWQ)
to ask questions, discuss your ideas, and connect with other contributors.

## Code of Conduct

See our Code of Conduct [here](https://docs.celestia.org/community/coc).

[centralized]: <https://github.com/rollkit/centralized-sequencer>
[astria]: <https://github.com/rollkit/astria-sequencer>
[dummy]: https://github.com/rollkit/go-sequencing/blob/main/test/dummy.go
[proxy]: https://github.com/rollkit/go-sequencing/tree/main/proxy
[gfi]: https://github.com/rollkit/go-da/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22
