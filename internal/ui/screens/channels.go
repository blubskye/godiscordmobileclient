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
	"sort"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/blubskye/godiscordmobileclient/internal/models"
	"github.com/blubskye/godiscordmobileclient/internal/ui/theme"
)

// ChannelsScreen displays channels for a guild
type ChannelsScreen struct {
	theme           *theme.Theme
	cache           CacheInterface
	onChannelSelect func(channelID string)
	onBack          func()

	list           widget.List
	backButton     widget.Clickable
	channelButtons map[string]*widget.Clickable
}

// NewChannelsScreen creates a new channels screen
func NewChannelsScreen(theme *theme.Theme, cache CacheInterface, onSelect func(channelID string), onBack func()) *ChannelsScreen {
	return &ChannelsScreen{
		theme:           theme,
		cache:           cache,
		onChannelSelect: onSelect,
		onBack:          onBack,
		channelButtons:  make(map[string]*widget.Clickable),
		list: widget.List{
			List: layout.List{Axis: layout.Vertical},
		},
	}
}

// Layout renders the channels screen
func (s *ChannelsScreen) Layout(gtx layout.Context, guildID string) layout.Dimensions {
	guild := s.cache.GetGuild(guildID)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutHeader(gtx, guild)
		}),

		// Channel list
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return s.layoutChannelList(gtx, guildID)
		}),
	)
}

func (s *ChannelsScreen) layoutHeader(gtx layout.Context, guild *models.Guild) layout.Dimensions {
	// Handle back button
	if s.backButton.Clicked(gtx) {
		s.onBack()
	}

	return layout.Stack{}.Layout(gtx,
		// Background
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(unit.Dp(56))}
			paint.FillShape(gtx.Ops, s.theme.BackgroundDark, clip.Rect{Max: size}.Op())
			return layout.Dimensions{Size: size}
		}),
		// Content
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

					// Guild name
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						name := ""
						if guild != nil {
							name = guild.Name
						}
						title := material.H6(s.theme.Theme, name)
						title.Color = s.theme.TextPrimary
						return title.Layout(gtx)
					}),
				)
			})
		}),
	)
}

func (s *ChannelsScreen) layoutChannelList(gtx layout.Context, guildID string) layout.Dimensions {
	channels := s.cache.GetGuildChannels(guildID)

	if len(channels) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(s.theme.Theme, "Loading channels...")
			label.Color = s.theme.TextSecondary
			return label.Layout(gtx)
		})
	}

	// Sort and organize channels
	organized := s.organizeChannels(channels)

	return material.List(s.theme.Theme, &s.list).Layout(gtx, len(organized), func(gtx layout.Context, i int) layout.Dimensions {
		item := organized[i]
		if item.isCategory {
			return s.layoutCategory(gtx, item.channel)
		}
		return s.layoutChannelItem(gtx, item.channel)
	})
}

type channelItem struct {
	channel    *models.Channel
	isCategory bool
}

func (s *ChannelsScreen) organizeChannels(channels []*models.Channel) []channelItem {
	// Separate categories and other channels
	categories := make(map[string]*models.Channel)
	categorized := make(map[string][]*models.Channel)
	uncategorized := []*models.Channel{}

	for _, ch := range channels {
		if ch.Type == models.ChannelTypeGuildCategory {
			categories[ch.ID] = ch
		} else if ch.ParentID != "" {
			categorized[ch.ParentID] = append(categorized[ch.ParentID], ch)
		} else {
			uncategorized = append(uncategorized, ch)
		}
	}

	// Sort categories by position
	var catList []*models.Channel
	for _, cat := range categories {
		catList = append(catList, cat)
	}
	sort.Slice(catList, func(i, j int) bool {
		return catList[i].Position < catList[j].Position
	})

	// Build organized list
	var result []channelItem

	// Add uncategorized first
	sort.Slice(uncategorized, func(i, j int) bool {
		return uncategorized[i].Position < uncategorized[j].Position
	})
	for _, ch := range uncategorized {
		if ch.IsText() || ch.IsVoice() {
			result = append(result, channelItem{channel: ch, isCategory: false})
		}
	}

	// Add categories with their channels
	for _, cat := range catList {
		result = append(result, channelItem{channel: cat, isCategory: true})

		children := categorized[cat.ID]
		sort.Slice(children, func(i, j int) bool {
			return children[i].Position < children[j].Position
		})
		for _, ch := range children {
			if ch.IsText() || ch.IsVoice() {
				result = append(result, channelItem{channel: ch, isCategory: false})
			}
		}
	}

	return result
}

func (s *ChannelsScreen) layoutCategory(gtx layout.Context, category *models.Channel) layout.Dimensions {
	return layout.Inset{
		Left:   unit.Dp(16),
		Right:  unit.Dp(16),
		Top:    unit.Dp(16),
		Bottom: unit.Dp(4),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := material.Caption(s.theme.Theme, category.Name)
		label.Color = s.theme.TextMuted
		return label.Layout(gtx)
	})
}

func (s *ChannelsScreen) layoutChannelItem(gtx layout.Context, channel *models.Channel) layout.Dimensions {
	// Get or create button
	btn, ok := s.channelButtons[channel.ID]
	if !ok {
		btn = &widget.Clickable{}
		s.channelButtons[channel.ID] = btn
	}

	// Only allow clicking text channels
	if btn.Clicked(gtx) && channel.IsText() {
		s.onChannelSelect(channel.ID)
	}

	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Left:   unit.Dp(8),
			Right:  unit.Dp(8),
			Top:    unit.Dp(2),
			Bottom: unit.Dp(2),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return s.layoutChannelItemContent(gtx, channel, btn.Hovered())
		})
	})
}

func (s *ChannelsScreen) layoutChannelItemContent(gtx layout.Context, channel *models.Channel, hovered bool) layout.Dimensions {
	bgColor := s.theme.Background
	if hovered {
		bgColor = theme.ColorBackgroundHover
	}

	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(unit.Dp(36))}
			rr := clip.RRect{
				Rect: image.Rectangle{Max: size},
				NE:   4, NW: 4, SE: 4, SW: 4,
			}
			paint.FillShape(gtx.Ops, bgColor, rr.Op(gtx.Ops))
			return layout.Dimensions{Size: size}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Left:   unit.Dp(12),
				Right:  unit.Dp(12),
				Top:    unit.Dp(8),
				Bottom: unit.Dp(8),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					// Channel icon
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						icon := "#"
						if channel.IsVoice() {
							icon = "🔊"
						}
						label := material.Body2(s.theme.Theme, icon)
						label.Color = s.theme.TextMuted
						return label.Layout(gtx)
					}),

					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),

					// Channel name
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						name := material.Body2(s.theme.Theme, channel.Name)
						name.Color = s.theme.TextSecondary
						if hovered {
							name.Color = s.theme.TextPrimary
						}
						return name.Layout(gtx)
					}),
				)
			})
		}),
	)
}
