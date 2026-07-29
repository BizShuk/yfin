# Hard Decisions

`docs/memory/` stores durable decisions that future work must preserve, especially decisions whose rationale is not obvious from code alone.

It does not store implementation plans. Active plans live at [plans/](../../plans/). Obsolete plans are deleted; completed plans with durable specification value move to [docs/specs/](../specs/).

Each memory document must state the decision, rationale, constraints, and the condition that would justify revisiting it. Current architecture and business scope remain defined by [CLAUDE.md](../../CLAUDE.md) and [README.md](../../README.md).

Current decisions:

- [Downstream cache ownership](2026-07-29-downstream-cache-ownership.md)
