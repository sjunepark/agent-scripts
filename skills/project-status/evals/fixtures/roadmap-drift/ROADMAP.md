# Roadmap

## Current

CSV report output

Status: in progress

Current state: The CSV formatter exists but the command does not route report
requests to it and has no integration test.

Next action: Wire `renderReport(records, "csv")` to the CSV formatter and add
an integration test.

## Next

Document the CSV output option for users.
