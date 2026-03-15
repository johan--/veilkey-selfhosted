# AGPL Compliance Checklist

This repository is licensed under `AGPL-3.0`.

This checklist is an operational aid for `veilkey-proxy`. It is not legal advice. Review it before exposing a modified build to remote users or distributing binaries, containers, or appliance images.

## Immediate Requirements

1. Publish the complete corresponding source for the exact deployed version.
2. Give remote users a clear way to obtain that source.
3. Keep copyright, license, and modification notices intact.
4. Ensure deployment artifacts point to the same source revision.

## Remote Network Use

Use this section when a modified `veilkey-proxy` instance is reachable by users over a network.

- Provide a prominent source link or source notice to remote users.
- Make the link point to the exact corresponding source for the running build.
- Keep the source available for as long as that build is offered to users.
- Include build identifiers so operators can map a running service to a source commit quickly.

Recommended implementation for this repo:

- Add a `Source` link or AGPL notice to any operator-facing landing page, terminal banner, proxy splash page, or admin UI.
- If there is no UI, provide the notice through the normal connection path users see first.
- Publish the source at a stable URL that includes the exact commit or release tag.

## Source Completeness

The published source should include everything needed to modify and rebuild the AGPL-covered work, including:

- application source
- build scripts
- deployment scripts used for the covered work
- interface definitions and config templates needed to build or install it

Do not rely on a partial mirror or a stale branch.

## Modification Notices

For modified versions:

- keep existing copyright and license notices
- state that the program was modified
- include a relevant modification date where appropriate
- avoid shipping binaries or images that omit the AGPL text

## Distribution Cases

Use this section if you distribute binaries, containers, VM images, or appliances.

- Ship the AGPL license text with the artifact.
- Provide equivalent access to corresponding source.
- If source is not embedded, provide a clear written offer or URL next to the artifact.
- Keep the source available long enough to satisfy AGPL obligations for that distribution path.

## Repo-Specific Checklist

- `LICENSE` matches `AGPL-3.0`.
- `README.md` states `AGPL-3.0`.
- Deployed host scripts come from a committed revision.
- The running host can report the current commit or release identifier.
- The operator path used by remote users includes a source notice.
- Mirrors such as `/opt/veilkey/veilkey-proxy` are kept in sync with the published source.

## Suggested Minimum Operator Workflow

1. Commit changes before deployment.
2. Push the exact commit to the published remote.
3. Deploy only from that pushed commit.
4. Record the deployed commit in the service banner, status output, or release notes.
5. Verify the source link presented to remote users resolves to that exact revision.

## Official References

- GNU AGPL v3 text: https://www.gnu.org/licenses/agpl-3.0.en.html
- GNU GPL/AGPL FAQ: https://www.gnu.org/licenses/gpl-faq.html
- GNU how-to for applying licenses: https://www.gnu.org/licenses/gpl-howto.html
