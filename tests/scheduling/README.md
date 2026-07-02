# Scheduling Tests

## Status

Scheduling tests are planned for post-MVP when the ComputeWheel CRD and player-piano scheduler are implemented.

## Target Tests

- Time slice rotation assigns correct snapshot to each Compute Context
- Pre-provisioning activates next domino container before current completes
- Node affinity routes dominos to the node owning snapshot data
- Compute Wheel rotation does not interrupt in-flight domino chains

## Current MVP

The MVP executes domino chains sequentially via CLI. Scheduling logic will be added with the Kubernetes controller-runtime reconciler.
