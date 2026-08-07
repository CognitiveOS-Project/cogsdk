/*
Package daemon provides the typed client for the cognitiveosd IPC surface.

It is the canonical home of the Unix-socket JSON envelope (Type/From/Payload)
and the system codes defined in cognitiveosd-api.md. Consumers that currently
define their own private variant (e.g. cpm/internal/daemon) should import this
package instead.
*/
package daemon
