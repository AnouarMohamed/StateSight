# Product

## Register

product

## Users

StateSight is used by platform engineers, SRE teams, and reliability or security stakeholders reviewing Kubernetes configuration drift under operational pressure. They need to scan quickly, inspect evidence carefully, and distinguish observations from conclusions before deciding what to do next.

## Product Purpose

StateSight compares desired state from Git with live cluster state, groups drift into incidents, and exposes provenance that supports investigation. The interface succeeds when an operator can identify affected workloads, understand the observed difference, judge the trustworthiness of its evidence, and take a deliberate next action without the system overstating causality.

## Brand Personality

Precise, calm, forensic. The product should have the density and familiar control language of a serious engineering tool, with a GitHub dark-interface sensibility: graphite work surfaces, clear borders, compact tables, and restrained state color.

## Anti-references

- Generic SaaS dashboard compositions with hero metrics, marketing copy, large decorative cards, or presentation-style panels.
- Purple or blue gradients, glow effects, glass surfaces, and visual decoration disconnected from operator work.
- Interfaces that bury provenance behind vague labels or imply attribution that the evidence does not support.
- Loose spacing, oversized typography, or mobile behavior that turns investigation data into an unreadable desktop table.

## Design Principles

1. Evidence before inference: observed values and provenance are primary; unobserved attribution is never implied.
2. Density with clarity: tables, tabs, and compact status surfaces support scanning without sacrificing readable hierarchy.
3. Familiar controls: standard navigation, buttons, filtering, badges, and forms should feel immediately reliable to engineering operators.
4. Restrained semantics: neutral surfaces carry most of the interface; color communicates selection, severity, trust, and command priority.
5. Resilient investigation: long resource paths, empty states, failed requests, keyboard operation, and narrow viewports must remain usable.

## Accessibility & Inclusion

Target WCAG AA contrast for product text and interactive states. Provide visible keyboard focus, semantic landmarks and form labels, status text in addition to color, touch-usable mobile controls, and reduced-motion behavior for users who request it.
