package utils

import (
	"image/color"
	"strconv"
	"strings"
)

func HexToNRGBA(hex string) color.NRGBA {
	hex = strings.TrimPrefix(hex, "#")
	var r, g, b, a uint8 = 0, 0, 0, 255
	if len(hex) == 6 {
		r64, _ := strconv.ParseUint(hex[0:2], 16, 8)
		g64, _ := strconv.ParseUint(hex[2:4], 16, 8)
		b64, _ := strconv.ParseUint(hex[4:6], 16, 8)
		r, g, b = uint8(r64), uint8(g64), uint8(b64)
	} else if len(hex) == 8 {
		r64, _ := strconv.ParseUint(hex[0:2], 16, 8)
		g64, _ := strconv.ParseUint(hex[2:4], 16, 8)
		b64, _ := strconv.ParseUint(hex[4:6], 16, 8)
		a64, _ := strconv.ParseUint(hex[6:8], 16, 8)
		r, g, b, a = uint8(r64), uint8(g64), uint8(b64), uint8(a64)
	}
	return color.NRGBA{R: r, G: g, B: b, A: a}
}

func InterpolateColor(startColor, endColor color.NRGBA, t float64) color.NRGBA {
	return color.NRGBA{
		R: uint8(float64(startColor.R)*(1-t) + float64(endColor.R)*t),
		G: uint8(float64(startColor.G)*(1-t) + float64(endColor.G)*t),
		B: uint8(float64(startColor.B)*(1-t) + float64(endColor.B)*t),
		A: uint8(float64(startColor.A)*(1-t) + float64(endColor.A)*t),
	}
}
