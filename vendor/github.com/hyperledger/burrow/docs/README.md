# Hyperledger Burrow Documentation

Hyperledger Burrow is a permissioned Ethereum smart-contract blockchain node. It executes Ethereum EVM smart contract code (usually written in [Solidity](https://solidity.readthedocs.io)) on a permissioned virtual machine. Burrow provides transaction finality and high transaction throughput on a proof-of-stake [Tendermint](https://tendermint.com) consensus engine.

![burrow logo](assets/images/burrow.png)

1. [Installation](INSTALL.md)
1. [Logging](LOGGING.md)
1. [Quickstart](quickstart)
  * [Single full node](quickstart/single-full-node.md) - start your first chain
  * [Send transactions](quickstart/send-transactions.md) - how to communicate with your Burrow chain
  * [Deploy contracts](quickstart/deploy-contracts.md) - interact with the Ethereum Virtual Machine
  * [Multiple validators](quickstart/multiple-validators.md) - advanced consensus setup
  * [Seed nodes](quickstart/seed-nodes.md) - add new node dynamically
  * [Kubernetes](https://github.com/helm/charts/tree/master/stable/burrow) - bootstraps a burrow network on a Kubernetes cluster

