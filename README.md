# Google Cloud Auth Proxy

[![latest tag](https://ghcr-badge.egpl.dev/rinx/gcloud-auth-proxy/latest_tag?trim=major&label=latest)](https://github.com/users/rinx/packages/container/package/gcloud-auth-proxy)
[![image size](https://ghcr-badge.egpl.dev/rinx/gcloud-auth-proxy/size)](https://github.com/users/rinx/packages/container/package/gcloud-auth-proxy)

This is a software that provides [Google-signed authentication tokens][google-token] in several ways.

Currently, it supports the following token types.

- [Google-signed ID Token][google-id-token]

[google-token]: https://cloud.google.com/docs/authentication/token-types
[google-id-token]: https://cloud.google.com/docs/authentication/get-id-token#go

## Usecase

TBW

## Usage

Deploy gcloud-auth-proxy as a sidecar with your application.

Container is available on `ghcr.io/rinx/gcloud-auth-proxy:latest`.

CLI flags are the followings.

```bash
Usage:
  gcloud-auth-proxy [flags]

Flags:
      --audience string                      default audience (required)
  -h, --help                                 help for gcloud-auth-proxy
      --host string                          server host (default "0.0.0.0")
      --port string                          server port (default "8100")
      --token-source-cache-duration string   token source cache duration (default "30m")
  -v, --version                              version for gcloud-auth-proxy
```

### Endpoints

| endpoint         | method | description      |
|------------------|--------|------------------|
| `/idtoken`       | POST   | returns ID Token |
| `/idtoken/proxy` |        | forwards HTTP request and appends ID Token to its header |
| `/healthz`       |        | health check endpoint |
| `/readyz`        |        | readiness check endpoint |

## Similar Projects

- [DazWilkin/gcp-oidc-token-proxy](https://github.com/DazWilkin/gcp-oidc-token-proxy)
