% podman-artifact-inspect 1

## NAME
podman\-artifact\-inspect - Inspect an OCI artifact

## SYNOPSIS
**podman artifact inspect** [*options*] [*name*] ...

## DESCRIPTION

Inspect one or more virtual machines

Obtain greater detail about Podman virtual machines. More than one virtual machine can be
inspected at once.

The default machine name is `podman-machine-default`. If a machine name is not specified as an argument,
then `podman-machine-default` will be inspected.

Rootless only.

## OPTIONS
#### **--format**

Print results with a Go template.
<!--
| **Placeholder**     | **Description**                                                   |
| ------------------- |-------------------------------------------------------------------|
| .Created ...        | Time when artifact was added to the local store (string, ISO3601) |
-->


#### **--help**

Print usage statement.

#### **--remote**

Instead of inspecting an OCI artifact in the local store, inspect it on an image registry.

## EXAMPLES

Inspect an OCI image in the local store.
```
$ podman artifact inspect quay.io/myartifact/myml:latest
```

## SEE ALSO
**[podman(1)](podman.1.md)**, **[podman-artifact(1)](podman-artifact.1.md)**

## HISTORY
April 2024, Originally compiled by Brent Baude <bbaude@redhat.com>
