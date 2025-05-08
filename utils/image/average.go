package image

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

func SortColorsByLuminance(hexColors []string) ([]string, error) {
    type colorWithLuma struct {
        hex  string
        luma float64
    }

    var colors []colorWithLuma

    for _, hex := range hexColors {
        r, g, b, err := hexToRGB(hex)
        if err != nil {
            return nil, err
        }
        luma := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
        colors = append(colors, colorWithLuma{hex, luma})
    }

    sort.Slice(colors, func(i, j int) bool {
        return colors[i].luma > colors[j].luma
    })

    result := make([]string, len(colors))
    for i, c := range colors {
        result[i] = c.hex
    }

    return result, nil
}

func AverageHexColors(hexColors []string) (string, error) {
    var rSum, gSum, bSum float64

    for _, hex := range hexColors {
        r, g, b, err := hexToRGB(hex)
        if err != nil {
            return "", err
        }
        rSum += float64(r)
        gSum += float64(g)
        bSum += float64(b)
    }

    count := float64(len(hexColors))
    avgR := uint8(math.Round(rSum / count))
    avgG := uint8(math.Round(gSum / count))
    avgB := uint8(math.Round(bSum / count))

    return rgbToHex(avgR, avgG, avgB), nil
}

func hexToRGB(hex string) (r, g, b uint8, err error) {
    hex = strings.TrimPrefix(hex, "#")
    if len(hex) != 6 {
        return 0, 0, 0, fmt.Errorf("invalid hex color length")
    }

    r64, err := strconv.ParseUint(hex[0:2], 16, 8)
    if err != nil {
        return 0, 0, 0, err
    }
    g64, err := strconv.ParseUint(hex[2:4], 16, 8)
    if err != nil {
        return 0, 0, 0, err
    }
    b64, err := strconv.ParseUint(hex[4:6], 16, 8)
    if err != nil {
        return 0, 0, 0, err
    }

    return uint8(r64), uint8(g64), uint8(b64), nil
}

func rgbToHex(r, g, b uint8) string {
    return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}