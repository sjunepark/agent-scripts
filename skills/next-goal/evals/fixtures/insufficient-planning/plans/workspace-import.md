# Workspace Import

## Outcome

Users can safely import an existing workspace archive into the current
workspace.

## Current state

The repository has no import implementation. The plan does not decide whether
conflicting records should replace existing data, merge with it, or stop the
import. It also leaves the supported archive versions and compatibility policy
unsettled. These choices affect data-loss risk and user-visible behavior.

## Next action

Settle the conflict and compatibility policies before implementation.
