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
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

// FillBackground fills the entire context with a color
func FillBackground(gtx layout.Context, col color.NRGBA, size image.Point) {
	paint.FillShape(gtx.Ops, col, clip.Rect{Max: size}.Op())
}

// Rect draws a filled rectangle
func Rect(gtx layout.Context, size image.Point, col color.NRGBA) layout.Dimensions {
	paint.FillShape(gtx.Ops, col, clip.Rect{Max: size}.Op())
	return layout.Dimensions{Size: size}
}

// RoundedRect draws a rounded rectangle
func RoundedRect(gtx layout.Context, size image.Point, radius int, col color.NRGBA) layout.Dimensions {
	rr := clip.RRect{
		Rect: image.Rectangle{Max: size},
		NE:   radius, NW: radius, SE: radius, SW: radius,
	}
	paint.FillShape(gtx.Ops, col, rr.Op(gtx.Ops))
	return layout.Dimensions{Size: size}
}

// Circle draws a filled circle
func Circle(gtx layout.Context, radius int, col color.NRGBA) layout.Dimensions {
	size := image.Point{X: radius * 2, Y: radius * 2}
	rr := clip.RRect{
		Rect: image.Rectangle{Max: size},
		NE:   radius, NW: radius, SE: radius, SW: radius,
	}
	paint.FillShape(gtx.Ops, col, rr.Op(gtx.Ops))
	return layout.Dimensions{Size: size}
}

// Dp converts density-independent pixels to pixels
func Dp(gtx layout.Context, dp unit.Dp) int {
	return gtx.Dp(dp)
}

// Inset creates uniform insets
func Inset(dp unit.Dp) layout.Inset {
	return layout.UniformInset(dp)
}

// HInset creates horizontal insets
func HInset(dp unit.Dp) layout.Inset {
	return layout.Inset{Left: dp, Right: dp}
}

// VInset creates vertical insets
func VInset(dp unit.Dp) layout.Inset {
	return layout.Inset{Top: dp, Bottom: dp}
}
