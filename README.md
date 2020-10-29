# und Upgrade Manager

The `und` upgrade manager is a modification of the original [cosmosd](https://github.com/regen-network/cosmosd) tool developed
by Regen Network.

This upgrade helper tool has been forked and modified to work with `und` binaries,
which do not currently implement the `x.upgrade` module upon which the original
`cosmosd` (now `cosmovisor`) are built to work with.

(A planned future major upgrade of the `und` binaries will include integration
of the Cosmos SDK `x.upgrade` module, along with the SDK v0.40.x upgrade. This tool
is an interim solution until that upgrade occurs.)

This is a tiny little shim around Cosmos SDK binaries and allows for smooth 
and configurable management of upgrading
binaries as a live chain is upgraded, and can be used to simplify validator
devops while doing upgrades or to make syncing a full node for genesis
simple. The upgrade manager will monitor the stdout of the daemon to look 
for Committed state messages at certain heights indicating a pending or required upgrade 
and act appropriately.

## Arguments

`und_upgrader` is a shim around a native `und` binary. All arguments passed to the upgrade manager 
command will be passed to the current daemon binary (as a subprocess).
 It will return stdout and stderr of the subprocess as
it's own. Because of that, it cannot accept any command line arguments, nor
print anything to output (unless it dies before executing a binary).

Configuration will be passed in the following environment variables:

* `DAEMON_HOME` is the location where upgrade binaries should be kept (can
be `$HOME/.gaiad` or `$HOME/.xrnd`)
* `DAEMON_NAME` is the name of the binary itself (eg. `xrnd`, `gaiad`)
* `DAEMON_ALLOW_DOWNLOAD_BINARIES` (optional) if set to `on` will enable auto-downloading of new binaries
(for security reasons, this is intended for fullnodes rather than validators)
* `DAEMON_RESTART_AFTER_UPGRADE` (optional) if set to `on` it will restart a the sub-process with the same args
(but new binary) after a successful upgrade. By default, the manager dies afterwards and allows the supervisor
to restart it if needed. Note that this will not auto-restart the child if there was an error.

## Folder Layout

`$DAEMON_HOME/upgrade_manager` is expected to belong completely to the upgrade manager and subprocesses
constrolled by it. Under this folder, we will see the following:

```
- genesis
  - bin
    - $DAEMON_NAME
- upgrades
  - <name>
    - bin
      - $DAEMON_NAME
- current -> upgrades/foo, genesis, etc
- plan.json
```

Each version of the chain is stored under either `genesis` or `upgrades/<name>`, which holds `bin/$DAEMON_NAME`
along with any other needed files (maybe the cli client? maybe some dlls?). `current` is a symlink to the currently
active folder (so `current/bin/$DAEMON_NAME` is the binary)

`plan.json` is a simple JSON file containing an array of upgrades for particular heights,
for example:

```json
{
  "upgrades":[
    {
      "height": 200,
      "version": "1.4.8"
    },
    {
      "height": 300,
      "version": "1.4.9"
    }
  ]
}
```

The above plan indicates that `und v1.4.8` should be installed at block height 200, 
and `v1.4.9` at block height 300, and so on.

Note: the `<name>` after `upgrades` is the '`version`' of the upgrade as specified in `plan.json`.

Please note that `$DAEMON_HOME/upgrade_manager` just stores the *binaries* and associated *program code*.
The `upgrader` binary can be stored in any typical location (eg `/usr/local/bin`). The actual blockchain
program will store it's data under `$UND_HOME` etc, which is independent of the `$DAEMON_HOME`. You can
choose to export `UND_HOME=$DAEMON_HOME` and then end up with a configuation like the following, but this
is left as a choice to the admin for best directory layout.

```
- .gaiad
  - config
  - data
  - upgrade_manager
```

## Usage

Basic Usage:

* The admin is responsible for installing the `und_upgrader` and setting it as a eg. systemd service to auto-restart, along with proper environmental variables
* The admin is responsible for installing the `genesis` folder manually
* The admin is responsible for acquiring or defining the network agreed `plan.json`
* The upgrade manager will set the `current` link to point to `genesis` at first start (when no `current` link exists)
* The admin is (generally) responsible for installing the `upgrades/<name>` folders manually
* The upgrade manager handles switching over the binaries at the correct points, so the admin can prepare days in advance and relax at upgrade time

Note that chains that wish to support upgrades may package up a genesis upgrade manager tar file with this info, just as they
prepare the genesis binary tar file. In fact, they may offer a tar file will all upgrades up to current point for easy download
for those who wish to sync a fullnode from start.

The `DAEMON` specific code, like the tendermint config, the application db, syncing blocks, etc is done as normal.
The same eg. `GAIA_HOME` directives and command-line flags work, just the binary name is different.

Example:

```bash
export DAEMON_HOME=/path/to/.und_mainchain
export DAEMON_NAME=und
export DAEMON_RESTART_AFTER_UPGRADE=on
und_upgrader --home=/path/to/.und_mainchain
```

This will start the `und_upgrader`, passing all arguments to the underlying `und`
process, and automatically restarting the `und` binary after each upgrade.

## Upgradeable Binary Specification

In the basic version, the `und_upgrader` will read the stdout log messages
to determine when an upgrade is needed:

* The `und_upgrader` scans stdout for `Committed state  module=state height= `
messages, specifically for heights as defined in `plan.json`
* If an upgrade is needed, the current `und` subprocess will be killed, 
`current` will point to the new version directory, according to `plan.json`
and the new `und` binary version will be launched.

## Distribution

A pre-defined package, containing all requried binaries, and `plan.json`
 may be distributed via our `mainnet` repository, making it simple for new
nodes to join the network and seamlessly sync from genesis.
