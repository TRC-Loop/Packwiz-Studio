package theme

import (
	"fyne.io/fyne/v2"
	fynetheme "fyne.io/fyne/v2/theme"

	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
)

// Size maps Fyne's metric roles onto the spacing scale and type scale.
func (Studio) Size(name fyne.ThemeSizeName) float32 {
	if v, ok := sizes[name]; ok {
		return v
	}
	return fynetheme.DefaultTheme().Size(name)
}

var sizes = map[fyne.ThemeSizeName]float32{
	// Type scale.
	fynetheme.SizeNameCaptionText:    tokens.TextCaption,
	fynetheme.SizeNameText:           tokens.TextBody,
	fynetheme.SizeNameSubHeadingText: tokens.TextSubHeading,
	fynetheme.SizeNameHeadingText:    tokens.TextHeading,
	fynetheme.SizeNameLineSpacing:    tokens.LineSpacing,

	// Spacing.
	fynetheme.SizeNamePadding:      tokens.SpaceSM,
	fynetheme.SizeNameInnerPadding: tokens.SpaceMD,

	// Radii.
	fynetheme.SizeNameButtonRadius:    tokens.RadiusControl,
	fynetheme.SizeNameInputRadius:     tokens.RadiusControl,
	fynetheme.SizeNameSelectionRadius: tokens.RadiusControl,
	fynetheme.SizeNameCardRadius:      tokens.RadiusCard,
	fynetheme.SizeNameDialogRadius:    tokens.RadiusOverlay,
	fynetheme.SizeNameMenuRadius:      tokens.RadiusOverlay,
	fynetheme.SizeNamePopupRadius:     tokens.RadiusOverlay,

	// Lines and icons.
	fynetheme.SizeNameSeparatorThickness: tokens.HairlineThickness,
	fynetheme.SizeNameInputBorder:        tokens.HairlineThickness,
	fynetheme.SizeNameInlineIcon:         tokens.IconInline,
}
