# Project Overview

## What StateSight Is

StateSight is a GitOps forensic platform for Kubernetes.

It is designed to help teams understand drift between:

- desired state from Git
- live state in Kubernetes

## Who It Is For

- platform engineers
- SRE teams
- security and reliability stakeholders who need drift visibility

## What v1 Aims To Do

- connect source and cluster context
- capture desired and live snapshots
- surface drift incidents in a clear way
- provide practical recommendations (ignore, monitor, investigate, reconcile)

## Out of Scope at First

- auto-remediation
- full GitOps controller replacement
- deep multi-cluster automation complexity
- advanced ML-style analysis

## Current Stage

Baseline scaffold in progress with API, worker, web, migrations, and seed flow.
