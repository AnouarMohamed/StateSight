---
name: StateSight Operator Console
description: Dense, evidence-first drift investigation workspace for platform operators.
colors:
  canvas: "#0d1117"
  surface: "#161b22"
  surface-subtle: "#0f141a"
  header: "#010409"
  border: "#30363d"
  border-subtle: "#21262d"
  text: "#e6edf3"
  text-muted: "#8b949e"
  link: "#4493f8"
  link-surface: "#13243d"
  action: "#238636"
  action-hover: "#2ea043"
  success: "#3fb950"
  success-surface: "#12251a"
  warning: "#d29922"
  warning-surface: "#2b210d"
  danger: "#f85149"
  danger-surface: "#2d1516"
typography:
  headline:
    fontFamily: "-apple-system, BlinkMacSystemFont, Segoe UI, Noto Sans, Helvetica, Arial, sans-serif"
    fontSize: "24px"
    fontWeight: 600
    lineHeight: 1.25
    letterSpacing: "normal"
  body:
    fontFamily: "-apple-system, BlinkMacSystemFont, Segoe UI, Noto Sans, Helvetica, Arial, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: "normal"
  label:
    fontFamily: "-apple-system, BlinkMacSystemFont, Segoe UI, Noto Sans, Helvetica, Arial, sans-serif"
    fontSize: "12px"
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: "normal"
rounded:
  control: "6px"
spacing:
  compact: "8px"
  standard: "16px"
  section: "24px"
components:
  button-primary:
    backgroundColor: "{colors.action}"
    textColor: "{colors.text}"
    rounded: "{rounded.control}"
    padding: "8px 12px"
    height: "36px"
  button-primary-hover:
    backgroundColor: "{colors.action-hover}"
    textColor: "{colors.text}"
    rounded: "{rounded.control}"
    padding: "8px 12px"
    height: "36px"
  panel:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    rounded: "{rounded.control}"
    padding: "16px"
---

# Design System: StateSight Operator Console

## Overview

**Creative North Star: "The Graphite Evidence Desk"**

StateSight is a compact investigation surface, not a presentation dashboard. Its visual structure takes cues from mature engineering tools: graphite work areas, thin structural borders, dense tables, quiet labels, and controls located where decisions occur. Page composition prioritizes resources, drift fields, provenance records, and actions over explanatory marketing copy.

The default canvas is dark because operators may review incidents for sustained periods and because dense code-like identifiers remain legible against controlled contrast. Blue is reserved for navigation and links; green is reserved for commands or confirmed states; amber and red represent risk or attention. Nothing uses color alone as its meaning.

## Colors

The interface uses a restrained dark-neutral strategy. `canvas`, `surface`, `surface-subtle`, and `header` define hierarchy without decorative effects. `link` marks navigation and selected interactive state. `action` is used only for commands such as starting an analysis. Success, warning, and danger pairs provide readable text and tinted surfaces for status badges and observed drift values.

Borders are structural: `border` separates panels and controls, while `border-subtle` separates rows in dense data. Gradients, glows, translucent decoration, and saturated inactive states are outside this system.

## Typography

The product uses the native system sans stack to feel consistent with the operator's workstation and keep loading inexpensive. Headings remain compact at `24px` or below; data labels and badges use `12px`; core tables and controls use `14px`. Resource references, revisions, fields, actor identifiers, and numeric confidence values use the native monospace stack.

Long resource names and metadata must wrap or break inside their available width. Headings are not display typography, and body copy is kept short because this is a working surface rather than documentation.

## Elevation

Hierarchy is produced by tonal surfaces and one-pixel borders, not floating cards. Panels sit directly on the canvas and contain their own headers and divided rows. The header is a darker stable anchor for global navigation. Shadows are intentionally absent from investigation content so status color and data values remain the primary visual signals.

## Components

Primary actions are compact green buttons with an icon where the command benefits from recognition, such as `Analyze`. Secondary controls use the neutral panel surface and border. Keyboard focus uses a two-pixel blue outline outside the component.

Panels are reserved for bounded tools or data groups: inventory totals, application tables, incident differences, provenance, timelines, and rule administration. Tables become stacked rows on narrow screens when horizontal comparison is unnecessary; comparison and rule tables retain controlled horizontal scrolling when preserving columns is materially useful.

Provenance records distinguish `Captured`, `Ownership signal`, and `Untrusted` states in text and color. `not-attributed` is displayed as `No actor observed`. Managed-field ownership always carries the non-causality statement beside the record.

Loading states use quiet row-shaped skeletons. Error states are bounded, readable, and offer retry where a failed read can be repeated. Controls preserve a minimum touch height on coarse pointers and all motion is reduced when the user requests reduced motion.

## Do's and Don'ts

### Do:

- Do keep the interface dense and scan-friendly with thin dividers, short labels, and predictable table alignment.
- Do expose observed values and provenance trust before recommendations or actor interpretation.
- Do use blue for links and focus, green for operator commands or confirmed status, amber for caution, and red for incidents requiring attention.
- Do preserve keyboard access, readable focus, wrapped identifiers, and mobile task completion.

### Don't:

- Don't use hero metric cards, marketing copy, decorative illustration, gradients, glow effects, or glass surfaces.
- Don't nest cards inside cards or lift page sections as floating promotional tiles.
- Don't invent causality or present Kubernetes field ownership as proof of who created drift.
- Don't hide action state behind color, hover-only affordances, or clipped long identifiers.
