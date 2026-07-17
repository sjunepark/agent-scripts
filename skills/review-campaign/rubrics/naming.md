# Naming Pass Rubric

The test is at the point of use: does the call site read correctly without opening the definition? Naming is a design tool here, not cosmetics — a wrong name misleads every future reader and agent session.

## Check

- **False friends.** Names that promise less or more than they do: `get*` that mutates or fetches remotely, `check*` that throws, `ensure*` that silently creates, handlers named after the event but doing business logic. These outrank any number of merely-bland names.
- **One concept, one word.** The same domain concept must carry the same name across contracts, persistence, services, and UI (the same entity drifting between workspace vs project, run vs execution vs job). Divergence across a seam is a finding for the seam's area too. Build a tiny glossary in the area Map while reviewing.
- **Blob names.** `Manager`, `Util`, `Helper`, `Service` without a qualifier, `data`, `info`, `item` in exported positions — flag when the vagueness hides a real responsibility (often co-reported with a structure finding; put it where the fix is).
- **Convention consistency.** Symbol-kind conventions hold across the area: Svelte components/stores, Effect services/layers/tags, Go exported vs internal, file name matches primary export, paired contract/guard naming stays paired.
- **Booleans and polarity.** `is/has/should` prefixes, no negated names (`notReady`, `disableX = false`).
- **Vocabulary matches docs.** Code terms align with ARCHITECTURE.md and domain docs; if the docs say one thing and symbols another, one of them is wrong — name which.

## Do not flag

- Established repo-wide conventions you would merely restyle.
- Short names in tight scopes (`i`, `db`, `tx`).
- Private locals with mediocre names — nit at most, and only when truly misleading.

## Severity and tier

False friend on an exported symbol: major. Cross-seam concept divergence: major (it multiplies). Bland-but-honest names: nit. Renames ripple, so they are always triage; propose batch renames per concept, prioritizing exported and contract names. Nothing in this pass is auto except comments/strings that misname a symbol they describe.
