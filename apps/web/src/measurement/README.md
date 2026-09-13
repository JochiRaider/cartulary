# measurement/

[Source overview / parent](../README.md)

Deterministic browser fixtures for Network Flow grid load measurement.

These measurement-only entry points compose production grid components with
deterministic fixture data. Ordinary application startup is owned by the
[source browser entry](../main.tsx).

## Files

| File | Responsibility |
| --- | --- |
| [main.tsx](main.tsx) | Browser entry point that mounts the deterministic Network Flow grid measurement fixture. |
| [NetworkFlowGridLoadFixture.tsx](NetworkFlowGridLoadFixture.tsx) | Measurement-only deterministic supported-load fixture composed from the production Network Flow grid components. |
