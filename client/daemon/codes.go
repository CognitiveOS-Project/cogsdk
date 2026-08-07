package daemon

// System codes are the only human-override mechanism in CognitiveOS. They
// bypass the AI and are handled directly by the daemon/Raw Model. See
// product-specs/specs/system-codes.md.
const (
	CodeWake     = "wake"
	CodeIdle     = "idle"
	CodeSecurity = "security"
	CodeReset    = "reset"
	CodeUnlock   = "unlock"
)

// Origins accepted for system codes.
const (
	OriginKeyboard = "keyboard"
	OriginVoice    = "voice"
	OriginCLI      = "cli"
)

// ValidCode reports whether code is a known system code.
func ValidCode(code string) bool {
	switch code {
	case CodeWake, CodeIdle, CodeSecurity, CodeReset, CodeUnlock:
		return true
	default:
		return false
	}
}
