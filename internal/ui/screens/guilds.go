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

	"github.com/blubskye/godiscordmobileclient/internal/models"
	"github.com/blubskye/godiscordmobileclient/internal/ui/theme"
)

// CacheInterface defines the cache methods needed by screens
type CacheInterface interface {
	GetUser() *models.User
	GetGuild(id string) *models.Guild
	GetGuilds() []*models.Guild
	GetChannel(id string) *models.Channel
	GetGuildChannels(guildID string) []*models.Channel
	GetDMChannels() []*models.Channel
	GetMessages(channelID string) []*models.Message
}

// GuildsScreen displays the list of guilds (servers)
type GuildsScreen struct {
	theme         *theme.Theme
	cache         CacheInterface
	onGuildSelect func(guildID string)

	list         widget.List
	guildButtons map[string]*widget.Clickable
}

// NewGuildsScreen creates a new guilds screen
func NewGuildsScreen(theme *theme.Theme, cache CacheInterface, onSelect func(guildID string)) *GuildsScreen {
	return &GuildsScreen{
		theme:         theme,
		cache:         cache,
		onGuildSelect: onSelect,
		guildButtons:  make(map[string]*widget.Clickable),
		list: widget.List{
			List: layout.List{Axis: layout.Vertical},
		},
	}
}

// Layout renders the guilds screen
func (s *GuildsScreen) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutHeader(gtx)
		}),

		// Guild list
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return s.layoutGuildList(gtx)
		}),
	)
}

func (s *GuildsScreen) layoutHeader(gtx layout.Context) layout.Dimensions {
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
				Left:   unit.Dp(16),
				Right:  unit.Dp(16),
				Top:    unit.Dp(16),
				Bottom: unit.Dp(16),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				title := material.H6(s.theme.Theme, "Servers")
				title.Color = s.theme.TextPrimary
				return title.Layout(gtx)
			})
		}),
	)
}

func (s *GuildsScreen) layoutGuildList(gtx layout.Context) layout.Dimensions {
	guilds := s.cache.GetGuilds()

	// Show loading or empty state
	if len(guilds) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(s.theme.Theme, "Loading servers...")
			label.Color = s.theme.TextSecondary
			return label.Layout(gtx)
		})
	}

	return material.List(s.theme.Theme, &s.list).Layout(gtx, len(guilds), func(gtx layout.Context, i int) layout.Dimensions {
		guild := guilds[i]
		return s.layoutGuildItem(gtx, guild)
	})
}

func (s *GuildsScreen) layoutGuildItem(gtx layout.Context, guild *models.Guild) layout.Dimensions {
	// Get or create button for this guild
	btn, ok := s.guildButtons[guild.ID]
	if !ok {
		btn = &widget.Clickable{}
		s.guildButtons[guild.ID] = btn
	}

	// Handle click
	if btn.Clicked(gtx) {
		s.onGuildSelect(guild.ID)
	}

	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Left:   unit.Dp(12),
			Right:  unit.Dp(12),
			Top:    unit.Dp(6),
			Bottom: unit.Dp(6),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return s.layoutGuildItemContent(gtx, guild, btn.Hovered())
		})
	})
}

func (s *GuildsScreen) layoutGuildItemContent(gtx layout.Context, guild *models.Guild, hovered bool) layout.Dimensions {
	// Background with hover effect
	bgColor := s.theme.Background
	if hovered {
		bgColor = theme.ColorBackgroundHover
	}

	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(unit.Dp(56))}
			rr := clip.RRect{
				Rect: image.Rectangle{Max: size},
				NE:   8, NW: 8, SE: 8, SW: 8,
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
					// Guild icon
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.layoutGuildIcon(gtx, guild)
					}),

					layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),

					// Guild name and member count
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								name := material.Body1(s.theme.Theme, guild.Name)
								name.Color = s.theme.TextPrimary
								return name.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								var countText string
								if guild.MemberCount > 0 {
									countText = formatNumber(guild.MemberCount) + " members"
								} else if guild.ApproximateMemberCount > 0 {
									countText = formatNumber(guild.ApproximateMemberCount) + " members"
								}
								if countText == "" {
									return layout.Dimensions{}
								}
								count := material.Caption(s.theme.Theme, countText)
								count.Color = s.theme.TextMuted
								return count.Layout(gtx)
							}),
						)
					}),
				)
			})
		}),
	)
}

func (s *GuildsScreen) layoutGuildIcon(gtx layout.Context, guild *models.Guild) layout.Dimensions {
	size := gtx.Dp(unit.Dp(40))

	// Draw circle background
	rr := clip.RRect{
		Rect: image.Rectangle{Max: image.Point{X: size, Y: size}},
		NE:   size / 2, NW: size / 2, SE: size / 2, SW: size / 2,
	}

	// If no icon, show first letter
	if guild.Icon == "" {
		paint.FillShape(gtx.Ops, s.theme.Accent, rr.Op(gtx.Ops))

		// Center the first letter
		return layout.Stack{Alignment: layout.Center}.Layout(gtx,
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Point{X: size, Y: size}}
			}),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				initial := ""
				if len(guild.Name) > 0 {
					initial = string([]rune(guild.Name)[0])
				}
				label := material.Body1(s.theme.Theme, initial)
				label.Color = s.theme.TextPrimary
				return label.Layout(gtx)
			}),
		)
	}

	// TODO: Load actual guild icon image
	paint.FillShape(gtx.Ops, theme.ColorBackgroundLight, rr.Op(gtx.Ops))
	return layout.Dimensions{Size: image.Point{X: size, Y: size}}
}

// formatNumber formats a number with K/M suffixes
func formatNumber(n int) string {
	if n >= 1000000 {
		return itoa(n/1000000) + "M"
	}
	if n >= 1000 {
		return itoa(n/1000) + "K"
	}
	return itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
