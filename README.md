[![Stable](https://img.shields.io/badge/status-stable-brightgreen?style=for-the-badge)](https://github.com/kubewarden/community/blob/main/REPOSITORIES.md#stable)

# kubewarden-trusted-registry

`kubewarden-trusted-registry` is a Kubewarden policy implemented in Go, designed to ensure Kubernetes Pods only use container images from trusted registries.

## Introduction

This policy validates the registry source of container images in Kubernetes Pods against a user-defined list of trusted registries. If an image's registry is not in the trusted list, the request will be rejected.

Users can define the list of trusted registries through policy runtime configuration, for example:

```json
{
  "trusted_registries": ["docker-dev.xxx.xx.net", "docker-stg.xxx.xxx.net"]
}
```

### Features

- Supports image validation for multi-container Pods
- Provides user-friendly error messages indicating untrusted images
- Allows dynamic configuration of trusted registries through policy settings

## Code Structure

- `settings.go`: Handles policy configuration parsing and validation logic
- `validate.go`: Implements the actual validation logic to ensure Pod images meet requirements
- `main.go`: Entry point for policy registration
- `validate_test.go`: Contains unit tests and integration tests for the policy

## Implementation Details

> **Note**: WebAssembly is a rapidly evolving technology field. This project is based on the Go ecosystem as of 2023.

Since the official Go compiler currently cannot generate WebAssembly binaries that run outside browsers, this policy uses the [TinyGo](https://tinygo.org/) compiler for building.

For JSON data processing (such as policy configuration and Kubernetes requests), we use the following tools:

- [kubewarden/k8s-objects](https://github.com/kubewarden/k8s-objects): Provides Kubernetes type compatibility implementation for TinyGo
- [gjson](https://github.com/tidwall/gjson): Efficient library for fast JSON data querying
- [mapset](https://github.com/deckarep/golang-set): Generic implementation for set operations
- [kubewarden/policy-sdk-go](https://github.com/kubewarden/policy-sdk-go): Provides helper functions for policy development

Additionally, we strongly recommend using the latest version of the TinyGo compiler to avoid runtime errors caused by insufficient reflection support.

## Testing

### Unit Tests

Unit tests are implemented using the Go testing framework and defined in `_test.go` files. These tests can be run using the official Go compiler:

```
make test
```

### End-to-End Tests

End-to-end tests verify the actual behavior of the compiled WebAssembly module. These tests are implemented using [bats](https://github.com/bats-core/bats-core) and executed by loading the policy through the `kwctl` CLI:

```
make e2e-tests
```

## Automation

This project integrates the following [GitHub Actions](https://docs.github.com/en/actions):

- **`unit-tests`**: Runs Go unit tests
- **`e2e-tests`**: Builds WebAssembly policy, installs `bats`, and runs end-to-end tests
- **`release`**: Builds WebAssembly policy and pushes it to user-defined OCI registry (e.g., [ghcr](https://ghcr.io))

## How to Use

1. **Configure Policy**: Add trusted registry list to policy configuration:

   ```
   {
       "trusted_registries": ["trusted-registry.io", "secure-images.com"]
   }
   ```

2. **Deploy Policy**: Upload WebAssembly module to supported policy runtime environments, such as Kubewarden controller.

3. **Verify**: Submit Pod creation requests and observe policy validation behavior. If a container image's registry is not in the trusted list, the request will be rejected.
