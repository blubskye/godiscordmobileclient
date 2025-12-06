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

package screens

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/blubskye/godiscordmobileclient/internal/ui/theme"
)

const (
	AppName    = "Discord Mobile Client"
	AppVersion = "1.0.0"
	SourceURL  = "https://github.com/blubskye/godiscordmobileclient"
	License    = "GNU Affero General Public License v3.0"
)

// AboutScreen displays information about the application
type AboutScreen struct {
	theme      *theme.Theme
	onBack     func()
	backButton widget.Clickable
	sourceLink widget.Clickable
}

// NewAboutScreen creates a new about screen
func NewAboutScreen(theme *theme.Theme, onBack func()) *AboutScreen {
	return &AboutScreen{
		theme:  theme,
		onBack: onBack,
	}
}

// Layout renders the about screen
func (s *AboutScreen) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutHeader(gtx)
		}),

		// Content
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return s.layoutContent(gtx)
		}),
	)
}

func (s *AboutScreen) layoutHeader(gtx layout.Context) layout.Dimensions {
	if s.backButton.Clicked(gtx) {
		s.onBack()
	}

	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(unit.Dp(56))}
			paint.FillShape(gtx.Ops, s.theme.BackgroundDark, clip.Rect{Max: size}.Op())
			return layout.Dimensions{Size: size}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Left:   unit.Dp(8),
				Right:  unit.Dp(16),
				Top:    unit.Dp(8),
				Bottom: unit.Dp(8),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					// Back button
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.backButton.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Left:   unit.Dp(8),
								Right:  unit.Dp(8),
								Top:    unit.Dp(8),
								Bottom: unit.Dp(8),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								label := material.Body1(s.theme.Theme, "←")
								label.Color = s.theme.TextPrimary
								return label.Layout(gtx)
							})
						})
					}),

					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),

					// Title
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						title := material.H6(s.theme.Theme, "About")
						title.Color = s.theme.TextPrimary
						return title.Layout(gtx)
					}),
				)
			})
		}),
	)
}

func (s *AboutScreen) layoutContent(gtx layout.Context) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Left:  unit.Dp(32),
			Right: unit.Dp(32),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				// App icon placeholder
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					size := gtx.Dp(unit.Dp(80))
					rr := clip.RRect{
						Rect: image.Rectangle{Max: image.Point{X: size, Y: size}},
						NE:   16, NW: 16, SE: 16, SW: 16,
					}
					paint.FillShape(gtx.Ops, s.theme.Accent, rr.Op(gtx.Ops))

					return layout.Stack{Alignment: layout.Center}.Layout(gtx,
						layout.Stacked(func(gtx layout.Context) layout.Dimensions {
							return layout.Dimensions{Size: image.Point{X: size, Y: size}}
						}),
						layout.Stacked(func(gtx layout.Context) layout.Dimensions {
							label := material.H4(s.theme.Theme, "D")
							label.Color = s.theme.TextPrimary
							return label.Layout(gtx)
						}),
					)
				}),

				layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),

				// App name
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := material.H5(s.theme.Theme, AppName)
					title.Color = s.theme.TextPrimary
					title.Alignment = 1 // Center
					return title.Layout(gtx)
				}),

				layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),

				// Version
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					version := material.Body1(s.theme.Theme, "Version "+AppVersion)
					version.Color = s.theme.TextSecondary
					return version.Layout(gtx)
				}),

				layout.Rigid(layout.Spacer{Height: unit.Dp(32)}.Layout),

				// License
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Body2(s.theme.Theme, "Licensed under")
					label.Color = s.theme.TextMuted
					return label.Layout(gtx)
				}),

				layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),

				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					license := material.Body1(s.theme.Theme, License)
					license.Color = s.theme.TextPrimary
					return license.Layout(gtx)
				}),

				layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),

				// Source code link
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Body2(s.theme.Theme, "Source Code")
					label.Color = s.theme.TextMuted
					return label.Layout(gtx)
				}),

				layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),

				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return s.sourceLink.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						link := material.Body1(s.theme.Theme, SourceURL)
						link.Color = theme.ColorTextLink
						return link.Layout(gtx)
					})
				}),

				layout.Rigid(layout.Spacer{Height: unit.Dp(32)}.Layout),

				// Copyright
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					copyright := material.Caption(s.theme.Theme, "Copyright (C) 2025 blubskye")
					copyright.Color = s.theme.TextMuted
					return copyright.Layout(gtx)
				}),

				layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),

				// AGPL notice
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					notice := material.Caption(s.theme.Theme, "This is free software. You are free to change and redistribute it under the terms of the AGPL-3.0 license.")
					notice.Color = s.theme.TextMuted
					notice.Alignment = 1 // Center
					return notice.Layout(gtx)
				}),

				layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),

				// Disclaimer
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					disclaimer := material.Caption(s.theme.Theme, "This is an unofficial client. Not affiliated with Discord Inc.")
					disclaimer.Color = s.theme.TextMuted
					disclaimer.Alignment = 1 // Center
					return disclaimer.Layout(gtx)
				}),
			)
		})
	})
}
