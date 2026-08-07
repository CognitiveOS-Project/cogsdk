/*
Package cpm provides the typed CPM RPC client.

Per ADR-011, it is the only path to nested .cgp tools: tool invocation is
brokered by the cpm daemon (ADR-004 validated tool domains) and cogsdk ships
typed clients only — it never executes tools directly.
*/
package cpm
