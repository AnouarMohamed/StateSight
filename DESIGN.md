---
name: StateSight Operator Console
description: Dense, evidence-first drift investigation workspace for platform operators.
colors:
  canvas: "#0c0a09"
  surface: "#0c0a09"
  surface-subtle: "#131211"
  header: "#0c0a09"
  border: "#292524"
  border-subtle: "#1a1918"
  text: "#e7e5e4"
  text-muted: "#a8a29e"
  text-dim: "#8f8883"
  link: "#f97316"
  link-surface: "#28150d"
  action: "#ea580c"
  action-hover: "#f97316"
  success: "#4ade80"
  success-surface: "#101d15"
  warning: "#fbbf24"
  warning-surface: "#221a0c"
  danger: "#fb7185"
  danger-surface: "#241013"
typography:
  headline:
    fontFamily: "-apple-system, BlinkMacSystemFont, Segoe UI, Noto Sans, Helvetica, Arial, sans-serif"
    fontSize: "16px"
    fontWeight: 500
    lineHeight: 1.25
    letterSpacing: "0"
  body:
    fontFamily: "-apple-system, BlinkMacSystemFont, Segoe UI, Noto Sans, Helvetica, Arial, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: "0"
  label:
    fontFamily: "-apple-system, BlinkMacSystemFont, Segoe UI, Noto Sans, Helvetica, Arial, sans-serif"
    fontSize: "14px"
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: "0"
rounded:
  control: "0px"
spacing:
  compact: "8px"
  standard: "16px"
  section: "16px"
components:
  button-primary:
    backgroundColor: "{colors.action}"
    textColor: "{colors.canvas}"
    rounded: "{rounded.control}"
    padding: "8px 12px"
    height: "32px"
  button-primary-hover:
    backgroundColor: "{colors.action-hover}"
    textColor: "{colors.canvas}"
    rounded: "{rounded.control}"
    padding: "8px 12px"
    height: "32px"
  panel:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    rounded: "{rounded.control}"
    padding: "12px"
---

# Design System: StateSight Operator Console

## Overview

**Creative North Star: "The Git Evidence Terminal"**

StateSight is a compact investigation surface, not a presentation dashboard. Its visual language is adapted from the supplied GitOps forensics reference: warm-black working canvas, thin structural rules, dense tables, quiet uppercase labels, an orange command signal, and a persistent evidence-navigation rail. Page composition prioritizes resources, drift fields, provenance records, and supported actions.

The official orange Git logomark anchors the shell because desired-state investigation begins with Git material. It identifies that relationship, not product ownership or sponsorship. It is sourced from the Git logo downloads page, where the logo is credited to Jason Long under CC BY 3.0.

## Colors

The interface uses warm stone-black neutrals instead of graphite-blue surfaces. `canvas`, `surface`, and `header` intentionally form one uninterrupted work area; borders and the subtle hover surface establish structure. Orange is reserved for active navigation, links, focus and commands such as starting an analysis. Success, warning and danger states remain textual and bordered as well as colored.

Borders are structural: `border` separates panels and controls, while `border-subtle` separates rows in dense data. Gradients, glows, translucent decoration, and saturated inactive states are outside this system.

## Typography

The product uses the native system sans stack to feel consistent with the operator's workstation and keep loading inexpensive. Main headings remain compact at `16px`; navigational labels, table headers, badges, controls and table values have a `14px` minimum on every viewport. Density is produced through compact row height, thin borders and concise copy rather than unreadably small labels. Resource references, revisions, fields, actor identifiers, counts and confidence values use the native monospace stack.

Long resource names and metadata must wrap or break inside their available width. Headings are not display typography, body copy is kept short because this is a working surface rather than documentation, and letter spacing remains fixed at `0`.

## Elevation

Hierarchy is produced by one-pixel borders and slight hover fills, not floating cards. Panels sit directly on the shared canvas and contain their own headers and divided rows. The desktop sidebar is a stable investigation rail; mobile uses a compact top band and horizontal primary routes. Shadows are intentionally absent.

## Components

Primary actions are compact orange buttons with dark text and an icon where the command benefits from recognition, such as `Analyze`. Orange is not used to decorate data; it denotes selection or command intent. Secondary controls use the neutral surface and border. Keyboard focus uses a two-pixel orange outline outside the component.

Panels are reserved for bounded tools or data groups: inventory totals, application tables, incident differences, provenance, timelines, and rule administration. Tables become stacked rows on narrow screens when horizontal comparison is unnecessary; evidence comparison tables retain controlled horizontal scrolling when preserving columns is materially useful. Rule administration stacks on narrow screens so edit and confirmation controls remain directly operable.

Provenance records distinguish `Captured`, `Ownership signal`, and `Untrusted` states in text and color. `not-attributed` is displayed as `No actor observed`. Evidence fact labels stack above technical values so long actor identifiers retain usable width on narrow screens. Managed-field ownership always carries the non-causality statement beside the record.

The sidebar exposes only real overview data: incidents, queued or processing analysis jobs, and tracked applications. Reference-only concepts such as AI insights, computed health, remediation actions, commit history and cluster status are not rendered until StateSight owns corresponding API contracts.

Application subviews use a semantic tablist with arrow, Home and End keyboard operation. On narrow screens, all three modes occupy stable columns with their count below the full label, avoiding clipping at compact phone widths. Loading states expose busy status, errors announce alerts, and successful queued actions announce status without shifting page structure.

Application-owned ignore rules expose explicit edit, activation and delete commands. Delete is destructive command styling only at confirmation time, and the confirmation states that suppression audit history remains retained. Inherited workspace rules remain visible as context and display no mutation controls.

Controls preserve a minimum touch height on coarse pointers. Color feedback transitions use a restrained 150 ms ease-out curve; collapsing the sidebar changes immediately because layout dimensions are not animated. All motion is reduced when the user requests reduced motion.

## Do's and Don'ts

### Do:

- Do keep the interface dense and scan-friendly with thin dividers, short labels, and predictable table alignment.
- Do expose observed values and provenance trust before recommendations or actor interpretation.
- Do use orange for links, focus, current route and operator commands; green for confirmed status; amber for caution; red for incidents requiring attention.
- Do preserve keyboard access, readable focus, wrapped identifiers, and mobile task completion.
- Do keep visible text at `14px` or above across viewport sizes.

### Don't:

- Don't use hero metric cards, marketing copy, decorative illustration, gradients, glow effects, or glass surfaces.
- Don't nest cards inside cards or lift page sections as floating promotional tiles.
- Don't invent causality or present Kubernetes field ownership as proof of who created drift.
- Don't import Supabase-backed reference views, AI analyst labels, or buttons unless a real StateSight API behavior supports them.
- Don't animate layout dimensions or compress labels below the readability floor to manufacture density.
- Don't hide action state behind color, hover-only affordances, or clipped long identifiers.
