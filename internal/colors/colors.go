package colors

// ANSI color and text attribute codes
const (
	// --- Attributes ---
	Reset         = "\033[0m"
	Bold          = "\033[1m"
	Dim           = "\033[2m"
	Italic        = "\033[3m"
	Underline     = "\033[4m"
	Blink         = "\033[5m" // Often not supported or disabled
	Inverse       = "\033[7m" // Swaps foreground and background
	Hidden        = "\033[8m" // For password input
	Strikethrough = "\033[9m"

	// --- Foreground Colors ---
	FgBlack   = "\033[30m"
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
	FgWhite   = "\033[37m"

	// --- High-Intensity Foreground Colors ---
	FgHiBlack   = "\033[90m" // Often rendered as gray
	FgHiRed     = "\033[91m"
	FgHiGreen   = "\033[92m"
	FgHiYellow  = "\033[93m"
	FgHiBlue    = "\033[94m"
	FgHiMagenta = "\033[95m"
	FgHiCyan    = "\033[96m"
	FgHiWhite   = "\033[97m"

	// --- Background Colors ---
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"

	// --- High-Intensity Background Colors ---
	BgHiBlack   = "\033[100m" // Often rendered as gray
	BgHiRed     = "\033[101m"
	BgHiGreen   = "\033[102m"
	BgHiYellow  = "\033[103m"
	BgHiBlue    = "\033[104m"
	BgHiMagenta = "\033[105m"
	BgHiCyan    = "\033[106m"
	BgHiWhite   = "\033[107m"
)
