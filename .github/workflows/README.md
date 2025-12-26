# Disabled Workflows

The following workflows are disabled because `go-shared` is a Go library, not a deployable service:

- `deploy-dev.yml.disabled`
- `deploy-staging.yml.disabled`
- `deploy-production.yml.disabled`

## Why are these disabled?

`go-shared` is a shared Go library that is consumed as a module dependency by other services. It does not run as a standalone service and therefore does not need deployment workflows.

## How is this library used?

Services consume this library by adding it as a dependency in their `go.mod`:

```go
module github.com/vhvplatform/my-service

require (
    github.com/vhvplatform/go-shared v1.0.0
)
```

## Active Workflows

The following workflows remain active and are appropriate for a Go library:

- **test.yml** - Runs tests on every push and PR
- **ci.yml** - Runs comprehensive CI checks (lint, test, build)
- **release.yml** - Creates releases when tags are pushed

## For Services

If you need deployment workflows, see the `go-infrastructure` repository for service deployment templates.
