# fxredis
[![Go Reference](https://pkg.go.dev/badge/github.com/fxlib/fxredis.svg)](https://pkg.go.dev/github.com/fxlib/fxredis)
![Tests](https://github.com/fxlib/fxredis/actions/workflows/go.yml/badge.svg)

Opinionated Redis components for building services powered by the Fx dependency injection framework. It also
assumes the use of Uber's zap logging framework and the 12-factor app practice of storing the configuration
in the environment. Specifically using [caarlo0/env](https://github.com/caarlos0/env).
