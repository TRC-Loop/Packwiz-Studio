package tokens

// Type scale. Four steps only, because more invites inconsistency.
const (
	// TextCaption is metadata: versions, paths, category chips.
	TextCaption float32 = 11
	// TextBody is the default size for labels, lists and inputs.
	TextBody float32 = 13
	// TextSubHeading titles a pane or a form section.
	TextSubHeading float32 = 15
	// TextHeading is the largest step: the launcher's app name and
	// a pack window's pack name.
	TextHeading float32 = 20

	// TextRailLabel names a section under its glyph in the activity rail.
	// It is below the caption step because the rail is narrow.
	TextRailLabel float32 = 10

	// LineSpacing is the gap between wrapped lines of text.
	LineSpacing float32 = 4
)

// Icon sizes.
const (
	// IconInline sits beside a line of body text.
	IconInline float32 = 16
	// IconActivity is a glyph in the activity bar rail.
	IconActivity float32 = 18
	// IconModCard is the thumbnail on a browser grid card.
	IconModCard float32 = 48
	// IconPackLogo is a pack's icon.png in the launcher and title area.
	IconPackLogo float32 = 32
)
