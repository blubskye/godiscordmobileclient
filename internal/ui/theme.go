// Copyright (C) 2025 blubskye
  //
  // This program is free software: you can redistribute it and/or modify
  // it under the terms of the GNU Affero General Public License as published by
  // the Free Software Foundation, either version 3 of the License, or
  // (at your option) any later version.
  //
  // This program is distributed in the hope that it will be useful,
  // but WITHOUT ANY WARRANTY; without even the implied warranty of
  // MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  // GNU Affero General Public License for more details.
  //
  // You should have received a copy of the GNU Affero General Public License
  // along with this program.  If not, see <https://www.gnu.org/licenses/>.
  //
  // Source code: https://github.com/blubskye/godiscordmobileclient

  
package ui

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/text"
	"gioui.org/widget/material"
)

// Discord color palette
var (
	// Background colors
	ColorBackground       = rgb(0x313338) // Main background
	ColorBackgroundDark   = rgb(0x1e1f22) // Sidebar background
	ColorBackgroundDarker = rgb(0x111214) // Server list background
	ColorBackgroundLight  = rgb(0x383a40) // Input background
	ColorBackgroundHover  = rgb(0x3d3f45) // Hover state

	// Text colors
	ColorTextPrimary   = rgb(0xf2f3f5) // Primary text
	ColorTextSecondary = rgb(0xb5bac1) // Secondary text
	ColorTextMuted     = rgb(0x949ba4) // Muted text
	ColorTextLink      = rgb(0x00a8fc) // Links

	// Brand colors
	ColorBlurple       = rgb(0x5865f2) // Discord blurple
	ColorBlurpleHover  = rgb(0x4752c4) // Blurple hover
	ColorGreen         = rgb(0x57f287) // Online/success
	ColorYellow        = rgb(0xfee75c) // Idle/warning
	ColorRed           = rgb(0xed4245) // DND/error
	ColorGray          = rgb(0x747f8d) // Offline

	// Status colors
	ColorOnline    = ColorGreen
	ColorIdle      = ColorYellow
	ColorDND       = ColorRed
	ColorOffline   = ColorGray
	ColorInvisible = ColorGray

	// UI elements
	ColorDivider    = rgb(0x3f4147)
	ColorScrollbar  = rgb(0x1e1f22)
	ColorMention    = rgba(0xfaa61a, 0x1a) // Mention background
	ColorSelection  = rgba(0x5865f2, 0x40) // Selection background
)

// rgb creates a color from a hex value
func rgb(hex uint32) color.NRGBA {
	return color.NRGBA{
		R: uint8((hex >> 16) & 0xff),
		G: uint8((hex >> 8) & 0xff),
		B: uint8(hex & 0xff),
		A: 0xff,
	}
}

// rgba creates a color with alpha from hex and alpha values
func rgba(hex uint32, alpha uint8) color.NRGBA {
	return color.NRGBA{
		R: uint8((hex >> 16) & 0xff),
		G: uint8((hex >> 8) & 0xff),
		B: uint8(hex & 0xff),
		A: alpha,
	}
}

// Theme holds the application theme
type Theme struct {
	*material.Theme

	// Colors
	Background       color.NRGBA
	BackgroundDark   color.NRGBA
	BackgroundLight  color.NRGBA
	TextPrimary      color.NRGBA
	TextSecondary    color.NRGBA
	TextMuted        color.NRGBA
	Accent           color.NRGBA
	AccentHover      color.NRGBA
	Online           color.NRGBA
	Idle             color.NRGBA
	DND              color.NRGBA
	Offline          color.NRGBA
	Divider          color.NRGBA
	Error            color.NRGBA
}

// NewTheme creates a Discord-like theme
func NewTheme() *Theme {
	th := material.NewTheme()

	// Configure fonts
	th.Shaper = text.NewShaper(text.WithCollection(defaultFonts()))
	th.Palette.Bg = ColorBackground
	th.Palette.Fg = ColorTextPrimary
	th.Palette.ContrastBg = ColorBlurple
	th.Palette.ContrastFg = ColorTextPrimary

	return &Theme{
		Theme:           th,
		Background:      ColorBackground,
		BackgroundDark:  ColorBackgroundDark,
		BackgroundLight: ColorBackgroundLight,
		TextPrimary:     ColorTextPrimary,
		TextSecondary:   ColorTextSecondary,
		TextMuted:       ColorTextMuted,
		Accent:          ColorBlurple,
		AccentHover:     ColorBlurpleHover,
		Online:          ColorOnline,
		Idle:            ColorIdle,
		DND:             ColorDND,
		Offline:         ColorOffline,
		Divider:         ColorDivider,
		Error:           ColorRed,
	}
}

// defaultFonts returns the default font collection
func defaultFonts() []font.FontFace {
	return []font.FontFace{
		{Font: font.Font{Typeface: "Go"}, Face: defaultFont()},
	}
}

// defaultFont returns a placeholder for the default font face
// In a real app, you'd load actual font files
func defaultFont() text.Face {
	return nil // Gio will use system fonts as fallback
}

// StatusColor returns the color for a user status
func (t *Theme) StatusColor(status string) color.NRGBA {
	switch status {
	case "online":
		return t.Online
	case "idle":
		return t.Idle
	case "dnd":
		return t.DND
	default:
		return t.Offline
	}
}

// RoleColor converts a Discord role color int to NRGBA
func RoleColor(colorInt int) color.NRGBA {
	if colorInt == 0 {
		return ColorTextPrimary
	}
	return color.NRGBA{
		R: uint8((colorInt >> 16) & 0xff),
		G: uint8((colorInt >> 8) & 0xff),
		B: uint8(colorInt & 0xff),
		A: 0xff,
	}
}

// WithAlpha returns a color with modified alpha
func WithAlpha(c color.NRGBA, alpha uint8) color.NRGBA {
	c.A = alpha
	return c
}

// Darken returns a darker version of the color
func Darken(c color.NRGBA, amount float32) color.NRGBA {
	return color.NRGBA{
		R: uint8(float32(c.R) * (1 - amount)),
		G: uint8(float32(c.G) * (1 - amount)),
		B: uint8(float32(c.B) * (1 - amount)),
		A: c.A,
	}
}

// Lighten returns a lighter version of the color
func Lighten(c color.NRGBA, amount float32) color.NRGBA {
	return color.NRGBA{
		R: uint8(float32(c.R) + (255-float32(c.R))*amount),
		G: uint8(float32(c.G) + (255-float32(c.G))*amount),
		B: uint8(float32(c.B) + (255-float32(c.B))*amount),
		A: c.A,
	}
}
