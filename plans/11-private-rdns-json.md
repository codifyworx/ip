# Private Reverse DNS JSON

**Issue:** #11
**Branch:** feat/dbrowning2-private-rdns-json
**Status:** ready-for-pr

## Objective

Move reverse DNS lookup out of documented public JSON and summary paths while keeping an undocumented JSON endpoint available for personal reverse-DNS use.

## Steps

- [x] Remove reverse DNS lookup from documented public paths.
- [x] Add an undocumented reverse-DNS JSON path.
- [x] Keep the hidden path out of README and rendered endpoint links.
- [x] Add focused route tests.
- [x] Validate formatting, tests, workflow syntax, and whitespace.

## Decisions

- Use `/json.rdns` for the undocumented personal endpoint.
- Do not manually deploy as part of this work.
