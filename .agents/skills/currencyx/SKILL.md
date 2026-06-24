---
name: currencyx
description: Work on OpenMeter currency primitives in pkg/currencyx for fiat and custom currency codes, rounding rules, calculators, allocation precision, fiat/custom boundaries, and callers in billing, charges, ledger, product catalog, subscriptions, or currency registry code.
---

# Currencyx

Use this skill when changes touch `pkg/currencyx` or any caller that depends on currency code shape, fiat/custom classification, calculator behavior, rounding, allocation, or invoice/ledger currency boundaries.

Also load the package skill for each touched area: `billing`, `charges`, `ledger`, `subscription`, `api`, `ent`, `db-migration`, and `test`.

## Boundary Model

- **Currency code**: durable identifier. ISO fiat codes and namespace-scoped custom codes share the same structural validation in `currencyx.Code`.
- **Currency interface**: shared behavior contract for modules that need `CurrencyCode`, `CurrencyType`, `Rounding`, `Calculator`, and `Validate`.
- **Fiat calculator**: uses the ISO definition for precision and preserves existing fiat rounding behavior.
- **Custom calculator**: uses configured custom rounding; default is whole-unit half-even bankers rounding.
- **Registry boundary**: owns custom currency definition, activation/archive rules, fiat-code collision checks, and future persisted rounding configuration.
- **Finance boundary**: snapshots fiat basis and applies fiat rounding when custom units become fiat amounts.
- **Ledger boundary**: records durable currency codes and balanced single-currency legs. Use calculator rounding only before posting when the upstream domain owns normalization.
- **Invoice boundary**: invoice currency stays fiat. Custom units must be materialized to fiat before billing invoice artifacts.

## Process

1. Name the surface before editing: code validation, rounding, calculator, allocation, registry, ledger fact, fiat materialization, or invoice boundary.
2. Keep `pkg/currencyx` free of imports from `openmeter/...`; callers can implement `currencyx.Currency` to supply configured custom rounding.
3. Do not add `ValidateFiat`, `IsFiat`, or similar split helpers. Use `CurrencyType()` at the boundary that truly requires fiat or custom.
4. Preserve fiat behavior unless the request explicitly changes fiat money rounding.
5. For custom currencies, default missing rounding config to half-even bankers rounding.
6. Keep allocation deterministic: use precision for units and largest-remainder distribution; do not silently preserve extra fractional custom units after rounding rules exist.
7. Verify with focused `pkg/currencyx` tests and caller package tests for every touched boundary.

## Review Checks

- Custom and fiat currencies share the `currencyx.Currency` interface.
- `Calculator.RoundToPrecision` applies the effective rounding rule.
- Invalid rounding precision or mode fails validation.
- Billing/invoice code rejects custom invoice currency explicitly instead of relying on `currencyx` to reject it.
- Ledger code accepts structurally valid custom codes and keeps transaction groups balanced per currency.
- Tests cover banker ties, configured custom precision, fiat regression behavior, invalid rounding config, and allocation precision.
