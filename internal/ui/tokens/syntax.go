package tokens

// Syntax colours for the file editor.
//
// The chrome stays grayscale, but a code editor is content rather than
// chrome: telling a key from a string from a comment at a glance is the
// whole point of highlighting, and four shades of grey do not do it. The
// hues here are muted enough to sit on ColorBG without glowing, and they
// are deliberately separate from the status colours above so that red
// still only ever means a failure.
var (
	// SyntaxComment is a hash comment.
	SyntaxComment = rgb(0x6A, 0x99, 0x55)
	// SyntaxTable is a table header such as [update.modrinth].
	SyntaxTable = rgb(0xD7, 0xBA, 0x7D)
	// SyntaxKey is a bare key on the left of an assignment.
	SyntaxKey = rgb(0x9C, 0xDC, 0xFE)
	// SyntaxString is a quoted value.
	SyntaxString = rgb(0xCE, 0x91, 0x78)
	// SyntaxNumber is a numeric or date literal.
	SyntaxNumber = rgb(0xB5, 0xCE, 0xA8)
	// SyntaxBool is true or false.
	SyntaxBool = rgb(0x56, 0x9C, 0xD6)
	// SyntaxPunct is structural: equals, brackets, braces, commas.
	SyntaxPunct = gray(0x80)
	// SyntaxKeyword is a reserved word.
	SyntaxKeyword = rgb(0xC5, 0x86, 0xC0)
	// SyntaxText is anything unclassified.
	SyntaxText = ColorText

	// SyntaxGutter is the line number column.
	SyntaxGutter = gray(0x60)
	// SyntaxGutterCurrent is the line number the caret is on.
	SyntaxGutterCurrent = ColorText
	// SyntaxCurrentLine tints the line the caret is on. It is barely above
	// the ground on purpose: it marks where you are without competing with
	// the selection.
	SyntaxCurrentLine = gray(0x27, 0x27, 0x2A)
)
