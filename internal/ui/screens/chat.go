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
	"context"
	"image"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/blubskye/godiscordmobileclient/internal/api"
	"github.com/blubskye/godiscordmobileclient/internal/models"
	"github.com/blubskye/godiscordmobileclient/internal/state"
	"github.com/blubskye/godiscordmobileclient/internal/ui/theme"
)

// ChatScreen displays messages for a channel
type ChatScreen struct {
	theme     *theme.Theme
	cache     *state.Cache
	apiClient *api.Client
	onBack    func()

	channelID string

	list         widget.List
	backButton   widget.Clickable
	sendButton   widget.Clickable
	messageInput widget.Editor
}

// NewChatScreen creates a new chat screen
func NewChatScreen(theme *theme.Theme, cache *state.Cache, apiClient *api.Client, onBack func()) *ChatScreen {
	return &ChatScreen{
		theme:     theme,
		cache:     cache,
		apiClient: apiClient,
		onBack:    onBack,
		list: widget.List{
			List: layout.List{
				Axis:        layout.Vertical,
				ScrollToEnd: true,
			},
		},
	}
}

// SetChannel sets the current channel
func (s *ChatScreen) SetChannel(channelID string) {
	s.channelID = channelID
}

// Layout renders the chat screen
func (s *ChatScreen) Layout(gtx layout.Context, channelID string) layout.Dimensions {
	s.channelID = channelID
	channel := s.cache.GetChannel(channelID)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutHeader(gtx, channel)
		}),

		// Messages
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return s.layoutMessages(gtx)
		}),

		// Input
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutInput(gtx, channel)
		}),
	)
}

func (s *ChatScreen) layoutHeader(gtx layout.Context, channel *models.Channel) layout.Dimensions {
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

					// Channel info
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								name := ""
								if channel != nil {
									name = "# " + channel.Name
								}
								title := material.Body1(s.theme.Theme, name)
								title.Color = s.theme.TextPrimary
								return title.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if channel == nil || channel.Topic == "" {
									return layout.Dimensions{}
								}
								topic := material.Caption(s.theme.Theme, channel.Topic)
								topic.Color = s.theme.TextMuted
								return topic.Layout(gtx)
							}),
						)
					}),
				)
			})
		}),
	)
}

func (s *ChatScreen) layoutMessages(gtx layout.Context) layout.Dimensions {
	messages := s.cache.GetMessages(s.channelID)

	if len(messages) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(s.theme.Theme, "No messages yet")
			label.Color = s.theme.TextSecondary
			return label.Layout(gtx)
		})
	}

	return material.List(s.theme.Theme, &s.list).Layout(gtx, len(messages), func(gtx layout.Context, i int) layout.Dimensions {
		msg := messages[i]
		var prevMsg *models.Message
		if i > 0 {
			prevMsg = messages[i-1]
		}
		return s.layoutMessage(gtx, msg, prevMsg)
	})
}

func (s *ChatScreen) layoutMessage(gtx layout.Context, msg *models.Message, prevMsg *models.Message) layout.Dimensions {
	// Check if this is a continuation of previous message
	isCompact := false
	if prevMsg != nil && prevMsg.Author != nil && msg.Author != nil {
		if prevMsg.Author.ID == msg.Author.ID {
			timeDiff := msg.Timestamp.Sub(prevMsg.Timestamp)
			if timeDiff < 5*time.Minute {
				isCompact = true
			}
		}
	}

	if isCompact {
		return s.layoutCompactMessage(gtx, msg)
	}
	return s.layoutFullMessage(gtx, msg)
}

func (s *ChatScreen) layoutFullMessage(gtx layout.Context, msg *models.Message) layout.Dimensions {
	return layout.Inset{
		Left:   unit.Dp(16),
		Right:  unit.Dp(16),
		Top:    unit.Dp(8),
		Bottom: unit.Dp(2),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx,
			// Avatar
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutAvatar(gtx, msg.Author)
			}),

			layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),

			// Message content
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					// Author and timestamp
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								name := "Unknown"
								if msg.Author != nil {
									name = msg.Author.DisplayName()
								}
								author := material.Body2(s.theme.Theme, name)
								author.Color = s.theme.TextPrimary
								author.Font.Weight = 600
								return author.Layout(gtx)
							}),

							layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),

							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								ts := formatTimestamp(msg.Timestamp)
								timestamp := material.Caption(s.theme.Theme, ts)
								timestamp.Color = s.theme.TextMuted
								return timestamp.Layout(gtx)
							}),
						)
					}),

					// Content
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						content := material.Body2(s.theme.Theme, msg.Content)
						content.Color = s.theme.TextSecondary
						return content.Layout(gtx)
					}),
				)
			}),
		)
	})
}

func (s *ChatScreen) layoutCompactMessage(gtx layout.Context, msg *models.Message) layout.Dimensions {
	return layout.Inset{
		Left:   unit.Dp(68), // Avatar width + spacing
		Right:  unit.Dp(16),
		Top:    unit.Dp(2),
		Bottom: unit.Dp(2),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		content := material.Body2(s.theme.Theme, msg.Content)
		content.Color = s.theme.TextSecondary
		return content.Layout(gtx)
	})
}

func (s *ChatScreen) layoutAvatar(gtx layout.Context, user *models.User) layout.Dimensions {
	size := gtx.Dp(unit.Dp(40))

	rr := clip.RRect{
		Rect: image.Rectangle{Max: image.Point{X: size, Y: size}},
		NE:   size / 2, NW: size / 2, SE: size / 2, SW: size / 2,
	}

	// Show colored circle with initial
	paint.FillShape(gtx.Ops, s.theme.Accent, rr.Op(gtx.Ops))

	return layout.Stack{Alignment: layout.Center}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Point{X: size, Y: size}}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			initial := "?"
			if user != nil && len(user.Username) > 0 {
				initial = string([]rune(user.Username)[0])
			}
			label := material.Body1(s.theme.Theme, initial)
			label.Color = s.theme.TextPrimary
			return label.Layout(gtx)
		}),
	)
}

func (s *ChatScreen) layoutInput(gtx layout.Context, channel *models.Channel) layout.Dimensions {
	// Handle send
	if s.sendButton.Clicked(gtx) {
		s.sendMessage()
	}

	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(unit.Dp(64))}
			paint.FillShape(gtx.Ops, s.theme.BackgroundDark, clip.Rect{Max: size}.Op())
			return layout.Dimensions{Size: size}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Left:   unit.Dp(16),
				Right:  unit.Dp(16),
				Top:    unit.Dp(12),
				Bottom: unit.Dp(12),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					// Input field
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return s.layoutInputField(gtx, channel)
					}),

					layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),

					// Send button
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.sendButton.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							size := gtx.Dp(unit.Dp(40))
							rr := clip.RRect{
								Rect: image.Rectangle{Max: image.Point{X: size, Y: size}},
								NE:   size / 2, NW: size / 2, SE: size / 2, SW: size / 2,
							}
							paint.FillShape(gtx.Ops, s.theme.Accent, rr.Op(gtx.Ops))

							return layout.Stack{Alignment: layout.Center}.Layout(gtx,
								layout.Stacked(func(gtx layout.Context) layout.Dimensions {
									return layout.Dimensions{Size: image.Point{X: size, Y: size}}
								}),
								layout.Stacked(func(gtx layout.Context) layout.Dimensions {
									label := material.Body1(s.theme.Theme, "→")
									label.Color = s.theme.TextPrimary
									return label.Layout(gtx)
								}),
							)
						})
					}),
				)
			})
		}),
	)
}

func (s *ChatScreen) layoutInputField(gtx layout.Context, channel *models.Channel) layout.Dimensions {
	hint := "Message"
	if channel != nil {
		hint = "Message #" + channel.Name
	}

	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(unit.Dp(40))}
			rr := clip.RRect{
				Rect: image.Rectangle{Max: size},
				NE:   20, NW: 20, SE: 20, SW: 20,
			}
			paint.FillShape(gtx.Ops, theme.ColorBackgroundLight, rr.Op(gtx.Ops))
			return layout.Dimensions{Size: size}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Left:   unit.Dp(16),
				Right:  unit.Dp(16),
				Top:    unit.Dp(10),
				Bottom: unit.Dp(10),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				s.messageInput.SingleLine = true
				ed := material.Editor(s.theme.Theme, &s.messageInput, hint)
				ed.Color = s.theme.TextPrimary
				ed.HintColor = s.theme.TextMuted
				return ed.Layout(gtx)
			})
		}),
	)
}

func (s *ChatScreen) sendMessage() {
	text := s.messageInput.Text()
	if text == "" {
		return
	}

	s.messageInput.SetText("")

	go func() {
		_, err := s.apiClient.SendMessage(context.Background(), s.channelID, text)
		if err != nil {
			// TODO: Show error to user
		}
	}()
}

func formatTimestamp(t time.Time) string {
	now := time.Now()
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return "Today at " + t.Format("3:04 PM")
	}
	yesterday := now.AddDate(0, 0, -1)
	if t.Year() == yesterday.Year() && t.YearDay() == yesterday.YearDay() {
		return "Yesterday at " + t.Format("3:04 PM")
	}
	return t.Format("01/02/2006 3:04 PM")
}
