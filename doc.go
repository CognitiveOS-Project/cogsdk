/*
Package cogsdk is the universal Go SDK for CognitiveOS.

It provides a single source of truth for the shared concerns that every
CognitiveOS component reimplements today: JSON framing, schema validation,
daemon and CPM IPC, and environment inspection.

Tier layout (see ADR-011):

  - client/ — implemented; typed clients for mcp, daemon, cpm, env
  - adk/    — staged; agent, intent, memory, prompt
  - cdk/    — staged; builder, cloudinit, target
*/
package cogsdk
