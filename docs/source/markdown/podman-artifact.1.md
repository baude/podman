% podman-artifact 1

## NAME
podman\-artifact - Manage OCI artifacts

## SYNOPSIS
**podman artifact** *subcommand*

## DESCRIPTION
`podman artifact` is a set of subcommands that manage OCI artifacts.

OCI artifacts are common way to distribute files that are associated with OCI images and
containers. Podman is capable of managing (pulling, inspecting, pushing) these artifacts
from its local "artifact store".

## SUBCOMMANDS

| Command | Man Page                                                   | Description             |
|---------|------------------------------------------------------------|-------------------------|
| inspect | [podman-artifact-inspect(1)](podman-artifact-inspect.1.md) | Inspect an OCI artifact |

## SEE ALSO
**[podman(1)](podman.1.md)**, **[podman-artifact-inspect(1)](podman-artifact-inspect.1.md)**

## HISTORY
Aug 2024, Originally compiled by Brent Baude <bbaude@redhat.com>
