<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking Protobuf, gRPC and REST routes used by end-users.
"CLI Breaking" for breaking CLI commands.
"API Breaking" for breaking exported APIs used by developers building on SDK.
Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]

### Features

* [#189](https://github.com/cosmos/rosetta/pull/189) Add system tests.

### Improvements

* [180](https://github.com/cosmos/rosetta/pull/180) Update to cosmos-sdk v0.52.0-rc.1.
* [180](https://github.com/cosmos/rosetta/pull/180) Added bech32 prefix flag.

### Bug Fixes

* [#190](https://github.com/cosmos/rosetta/pull/190) Fixed construction payloads to properly handle transaction signers.

## [v0.50.11](https://github.com/cosmos/rosetta/releases/tag/v0.50.11) 2024-12-19

### Improvements

* [#157](https://github.com/cosmos/rosetta/pull/157) Rosetta can now run without requiring a plugin.
* [174](https://github.com/cosmos/rosetta/pull/174) Update cosmos-sdk to v0.50.11.

### Bug Fixes

* [#157](https://github.com/cosmos/rosetta/pull/157) Added support for insecure connections to gRPC reflection servers.

## [v0.50.10](https://github.com/cosmos/rosetta/releases/tag/v0.50.10) 2024-09-25

### Improvements

* [155](https://github.com/cosmos/rosetta/pull/155) Update Go to 1.23.1 and cosmos-sdk to v0.50.10

## [v0.50.9](https://github.com/cosmos/rosetta/releases/tag/v0.50.9) 2024-08-07

### Improvements

* [141](https://github.com/cosmos/rosetta/pull/141) Update cosmos-sdk to v0.50.9

## [v0.50.8](https://github.com/cosmos/rosetta/releases/tag/v0.50.8) 2024-07-15

### Improvements

* [138](https://github.com/cosmos/rosetta/pull/138) Update cosmos-sdk to v0.50.8

## [v0.50.7](https://github.com/cosmos/rosetta/releases/tag/v0.50.7) 2024-06-05

### Improvements

* [128](https://github.com/cosmos/rosetta/pull/128) Update cosmos-sdk to v0.50.7

### Bug Fixes

* [123](https://github.com/cosmos/rosetta/pull/123) Correctly parse cometBFT finalize block.

## [v0.50.6](https://github.com/cosmos/rosetta/releases/tag/v0.50.6) 2024-04-23

### Improvements

* [114](https://github.com/cosmos/rosetta/pull/114) Update cosmos-sdk to v0.50.6

## [v0.50.5](https://github.com/cosmos/rosetta/releases/tag/v0.50.5) 2024-03-18

### Bug Fixes

* [99](https://github.com/cosmos/rosetta/pull/99) Rosetta now can connect to cometBFT over HTTPS.

### Improvements

* [93](https://github.com/cosmos/rosetta/pull/93) Removes the use of `LegacyMsg.GetSigners()` in favor of `codec.GetMsgV1Signers`.
* [#101](https://github.com/cosmos/rosetta/pull/101) Update to cosmos-sdk v0.50.5

## [v0.50.4](https://github.com/cosmos/rosetta/releases/tag/v0.50.4) 2024-02-26

### Improvements

* [#88](https://github.com/cosmos/rosetta/pull/88) Update to cosmos-sdk v0.50.4

## [v0.50.3+1](https://github.com/cosmos/rosetta/releases/tag/v0.50.3+1) 2024-01-07

> v0.50.3 has been retracted due to a mistake in dependencies. Please use v0.50.3+1 instead.

### Improvements

* [#73](https://github.com/cosmos/rosetta/pull/73) Update to cosmos-sdk v0.50.3
* [#70](https://github.com/cosmos/rosetta/pull/70) Coinbase accurate dockerfile.

### Bug Fixes

* [#82](https://github.com/cosmos/rosetta/pull/82) Fix cosmossdk.io/core dependencies.

## [v0.50.2](https://github.com/cosmos/rosetta/releases/tag/v0.50.2) 2023-12-12

### Improvements

* [#58](https://github.com/cosmos/rosetta/pull/58) Upgraded cosmos-sdk version and removed tip handling.
* [#37](https://github.com/cosmos/rosetta/pull/37) Dockerization of Rosetta.
* [#29](https://github.com/cosmos/rosetta/pull/29) Improvements on error handling.

## v0.47.x

* Migrated rosetta from cosmos-sdk repository to the standalone [repo](https://github.com/cosmos/rosetta).

### Improvements

* [#14272](https://github.com/cosmos/cosmos-sdk/pull/14272) Use `coinbase/rosetta-sdk-go/types` packages instead of comsos fork.

### Bug Fixes

* [#14285](https://github.com/cosmos/cosmos-sdk/pull/14285) Sets tendermint errors status codes to 500

## v0.2.0 2022-12-07

### Improvements

* [#14118](https://github.com/cosmos/cosmos-sdk/pull/14118) Allow rosetta to be installed as a standalone application.
* [#14061](https://github.com/cosmos/cosmos-sdk/pull/14061) Adds openapi specification.
* [#13832](https://github.com/cosmos/cosmos-sdk/pull/13832) Correctly populates rosetta's `/network/status` endpoint response. Rosetta's data api is divided into its own go files (account, block, mempool, network).

### Bug Fixes

* [#13832](https://github.com/cosmos/cosmos-sdk/pull/13832) Wrap tendermint RPC errors to rosetta errors.

## v0.1.0 2022-11-04

**From `v0.1.0` the minimum version of Tendermint is `v0.37+`, due event type changes.**

### Improvements

* [#13583](https://github.com/cosmos/cosmos-sdk/pull/13583) Extract rosetta to its own go.mod.
