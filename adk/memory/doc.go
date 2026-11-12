// Package memory provides memory persistence for CognitiveOS agents.
//
// Memory stores agent state, conversation history, and learned facts in a
// key-value store. The store is pluggable: an in-memory implementation is
// provided by default, and alternative backends can be plugged in via the
// Store interface.
package memory