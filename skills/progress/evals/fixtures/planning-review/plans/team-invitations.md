# Team Invitations

## Outcome

Workspace administrators can invite people with shareable invitation links.

## Current state

The proposal uses self-contained signed tokens so acceptance requires no stored
invitation record. A link remains valid until its workspace secret rotates and
the role encoded in the link replaces any role already held by the accepting
account.

## Next action

Add token generation and acceptance endpoints, then build the invitation form.
