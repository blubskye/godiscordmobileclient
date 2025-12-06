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
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/blubskye/godiscordmobileclient/internal/ui/theme"
)

// LoginScreen handles user authentication
type LoginScreen struct {
	theme    *theme.Theme
	onLogin  func(token string)

	// Login mode: token or email/password
	useToken bool

	// Token input
	tokenEditor widget.Editor

	// Email/password inputs
	emailEditor    widget.Editor
	passwordEditor widget.Editor
	mfaEditor      widget.Editor

	// Buttons
	loginButton     widget.Clickable
	switchModeButton widget.Clickable

	// State
	loading    bool
	error      string
	needsMFA   bool
	mfaTicket  string
}

// NewLoginScreen creates a new login screen
func NewLoginScreen(theme *theme.Theme, onLogin func(token string)) *LoginScreen {
	return &LoginScreen{
		theme:    theme,
		onLogin:  onLogin,
		useToken: true, // Default to token mode
	}
}

// Layout renders the login screen
func (s *LoginScreen) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return s.layoutContent(gtx)
			})
		}),
	)
}

func (s *LoginScreen) layoutContent(gtx layout.Context) layout.Dimensions {
	// Limit width for better appearance
	maxWidth := gtx.Dp(unit.Dp(350))
	if gtx.Constraints.Max.X > maxWidth {
		gtx.Constraints.Max.X = maxWidth
	}

	return layout.Inset{
		Left:  unit.Dp(24),
		Right: unit.Dp(24),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			// Title
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				title := material.H4(s.theme.Theme, "Discord Mobile")
				title.Color = s.theme.TextPrimary
				title.Alignment = 1 // Center
				return title.Layout(gtx)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),

			// Subtitle
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				sub := material.Body2(s.theme.Theme, "Unofficial client")
				sub.Color = s.theme.TextSecondary
				return sub.Layout(gtx)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(32)}.Layout),

			// Error message
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if s.error == "" {
					return layout.Dimensions{}
				}
				return s.layoutError(gtx)
			}),

			// Input fields
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if s.useToken {
					return s.layoutTokenInput(gtx)
				}
				return s.layoutEmailPasswordInput(gtx)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),

			// Login button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutLoginButton(gtx)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),

			// Switch mode button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutSwitchModeButton(gtx)
			}),
		)
	})
}

func (s *LoginScreen) layoutError(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Bottom: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Background
		size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(unit.Dp(40))}
		rr := clip.RRect{
			Rect: image.Rectangle{Max: size},
			NE:   4, NW: 4, SE: 4, SW: 4,
		}
		paint.FillShape(gtx.Ops, theme.WithAlpha(s.theme.Error, 40), rr.Op(gtx.Ops))

		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Body2(s.theme.Theme, s.error)
			label.Color = s.theme.Error
			return label.Layout(gtx)
		})
	})
}

func (s *LoginScreen) layoutTokenInput(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body2(s.theme.Theme, "Token")
			label.Color = s.theme.TextSecondary
			return label.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutEditor(gtx, &s.tokenEditor, "Paste your Discord token here", true)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			hint := material.Caption(s.theme.Theme, "Get your token from browser DevTools (F12) → Application → Local Storage → discord.com → token")
			hint.Color = s.theme.TextMuted
			return hint.Layout(gtx)
		}),
	)
}

func (s *LoginScreen) layoutEmailPasswordInput(gtx layout.Context) layout.Dimensions {
	if s.needsMFA {
		return s.layoutMFAInput(gtx)
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Email
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body2(s.theme.Theme, "Email")
			label.Color = s.theme.TextSecondary
			return label.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutEditor(gtx, &s.emailEditor, "Enter your email", false)
		}),

		layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),

		// Password
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body2(s.theme.Theme, "Password")
			label.Color = s.theme.TextSecondary
			return label.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutEditor(gtx, &s.passwordEditor, "Enter your password", true)
		}),
	)
}

func (s *LoginScreen) layoutMFAInput(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body2(s.theme.Theme, "Two-Factor Code")
			label.Color = s.theme.TextSecondary
			return label.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutEditor(gtx, &s.mfaEditor, "Enter 6-digit code", false)
		}),
	)
}

func (s *LoginScreen) layoutEditor(gtx layout.Context, editor *widget.Editor, hint string, mask bool) layout.Dimensions {
	// Background
	size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(unit.Dp(48))}
	rr := clip.RRect{
		Rect: image.Rectangle{Max: size},
		NE:   4, NW: 4, SE: 4, SW: 4,
	}
	paint.FillShape(gtx.Ops, theme.ColorBackgroundLight, rr.Op(gtx.Ops))

	return layout.Inset{
		Left:  unit.Dp(12),
		Right: unit.Dp(12),
		Top:   unit.Dp(12),
		Bottom: unit.Dp(12),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		editor.SingleLine = true
		if mask {
			editor.Mask = '•'
		}
		ed := material.Editor(s.theme.Theme, editor, hint)
		ed.Color = s.theme.TextPrimary
		ed.HintColor = s.theme.TextMuted
		return ed.Layout(gtx)
	})
}

func (s *LoginScreen) layoutLoginButton(gtx layout.Context) layout.Dimensions {
	// Handle click
	if s.loginButton.Clicked(gtx) && !s.loading {
		s.handleLogin()
	}

	// Button style
	btn := material.Button(s.theme.Theme, &s.loginButton, "Login")
	btn.Background = s.theme.Accent
	btn.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	if s.loading {
		btn.Text = "Connecting..."
		btn.Background = theme.Darken(s.theme.Accent, 0.3)
	}

	return btn.Layout(gtx)
}

func (s *LoginScreen) layoutSwitchModeButton(gtx layout.Context) layout.Dimensions {
	if s.switchModeButton.Clicked(gtx) {
		s.useToken = !s.useToken
		s.error = ""
		s.needsMFA = false
	}

	text := "Use email & password instead"
	if !s.useToken {
		text = "Use token instead"
	}

	btn := material.Button(s.theme.Theme, &s.switchModeButton, text)
	btn.Background = color.NRGBA{A: 0} // Transparent
	btn.Color = s.theme.TextSecondary

	return btn.Layout(gtx)
}

func (s *LoginScreen) handleLogin() {
	s.error = ""

	if s.useToken {
		token := s.tokenEditor.Text()
		if token == "" {
			s.error = "Please enter a token"
			return
		}
		s.onLogin(token)
	} else {
		// Email/password login would go here
		// For now, show a message that this feature is limited
		s.error = "Email login may be blocked by Discord. Use token instead."
	}
}
