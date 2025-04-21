# Heimdall

[![Go Report Card](https://goreportcard.com/badge/github.com/zenanetwork/iris)](https://goreportcard.com/report/github.com/zenanetwork/iris) [![CircleCI](https://circleci.com/gh/zenanetwork/iris/tree/master.svg?style=shield)](https://circleci.com/gh/zenanetwork/iris/tree/master) [![GolangCI Lint](https://github.com/zenanetwork/iris/actions/workflows/ci.yml/badge.svg)](https://github.com/zenanetwork/iris/actions)

Validator node for Matic Network. It uses peppermint, customized [Tendermint](https://github.com/tendermint/tendermint).

### Install from source

Make sure you have Go v1.20+ already installed.

### Install

```bash
$ make install
```

### Init Heimdall

```bash
$ irisd init
$ irisd init --chain=mainnet        Will init with genesis.json for mainnet
$ irisd init --chain=amoy           Will init with genesis.json for amoy
```

### Run Heimdall

```bash
$ irisd start
```

#### Usage

```bash
$ irisd start                       Will start for mainnet by default
$ irisd start --chain=mainnet       Will start for mainnet
$ irisd start --chain=amoy          Will start for amoy
$ irisd start --chain=local         Will start for local with NewSelectionAlgoHeight = 0
```

### Run rest server

```bash
$ irisd rest-server
```

### Run bridge

```bash
$ irisd bridge
```

### Develop using Docker

You can build and run Heimdall using the included Dockerfile in the root directory:

```bash
docker build -t iris .
docker run iris
```

### Documentation

Latest docs are [here](https://docs.polygon.technology/pos/).
