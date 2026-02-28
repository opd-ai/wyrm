# Wyrm

Procedurally generated first-person open-world RPG built in Go with Ebitengine.

## Directory Structure

```
cmd/client/              Client entry point (Ebitengine window)
cmd/server/              Authoritative game server entry point
config/                  Configuration loading (Viper)
pkg/engine/ecs/          Entity-Component-System core
pkg/engine/components/   ECS component definitions
pkg/engine/systems/      ECS system implementations
pkg/world/chunk/         World chunk management and streaming
pkg/rendering/raycast/   First-person raycasting renderer
pkg/rendering/texture/   Procedural texture generation
pkg/procgen/city/        Procedural city generation
pkg/audio/               Procedural audio synthesis
pkg/network/             Client-server networking
```

## Build

```
go build ./cmd/client
go build ./cmd/server
```

## Run

```
./client   # launches game window
./server   # starts authoritative server
```

## Configuration

Configuration is loaded from `config.yaml` in the working directory or `./config/` directory. Environment variables are also supported. Defaults are used when no config file is present.

See `config.yaml` for all available settings.

## Dependencies

- [Ebitengine v2](https://ebitengine.org/) — 2D game engine
- [Viper](https://github.com/spf13/viper) — configuration management
