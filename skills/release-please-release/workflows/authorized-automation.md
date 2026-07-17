# Authorized Automation Workflow

Use this workflow only when Release Please ownership is verified and the user has explicitly authorized the specific automation or release operation. Apply the version-confirmation gate in `SKILL.md` before every side-effecting step.

1. Wait for or inspect the relevant workflow and Release Please PR using the repository's documented tooling.
2. Review the PR before recommending merge:
   - exact version bump or per-component bumps;
   - changelog and release notes;
   - component scope;
   - breaking-change and migration notes;
   - CI status;
   - downstream workflows and publication effects triggered after merge.
3. Present the review summary and exact version or versions. If the user has not yet confirmed that version and the specific operation, stop and request confirmation.
4. Perform only the authorized operation. Do not broaden authorization to pushing, tagging, publishing, creating a release, merging another PR, or approving another irreversible step.
5. Wait for resulting checks or workflows when requested, then report the observed outcome, failures, remaining publication steps, and any operation that still needs separate authorization.
