package main

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
)

const dashboardURL = "http://127.0.0.1:18080"

var launchURL = openURL

func desktopMenuLabels() []string {
	return []string{"打开流量看板", "打开数据目录", "退出"}
}

func handleAlreadyRunning(alreadyRunning bool) error {
	if !alreadyRunning {
		return nil
	}
	return launchURL(dashboardURL)
}

func trayIconPNG() []byte {
	canvas := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	transparent := color.NRGBA{0, 0, 0, 0}
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			canvas.SetNRGBA(x, y, transparent)
		}
	}
	purple := color.NRGBA{95, 70, 255, 255}
	white := color.NRGBA{255, 255, 255, 255}
	for y := 2; y < 30; y++ {
		for x := 2; x < 30; x++ {
			dx, dy := x-16, y-16
			if dx*dx+dy*dy <= 14*14 {
				canvas.SetNRGBA(x, y, purple)
			}
		}
	}
	drawIconLine(canvas, 9, 20, 15, 12, white)
	drawIconLine(canvas, 15, 12, 23, 17, white)
	for _, point := range [][2]int{{9, 20}, {15, 12}, {23, 17}} {
		for y := point[1] - 2; y <= point[1]+2; y++ {
			for x := point[0] - 2; x <= point[0]+2; x++ {
				canvas.SetNRGBA(x, y, white)
			}
		}
	}
	var output bytes.Buffer
	_ = png.Encode(&output, canvas)
	return output.Bytes()
}

func drawIconLine(img *image.NRGBA, x0, y0, x1, y1 int, value color.NRGBA) {
	dx, sx := absInt(x1-x0), 1
	if x0 > x1 {
		sx = -1
	}
	dy, sy := -absInt(y1-y0), 1
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		img.SetNRGBA(x0, y0, value)
		img.SetNRGBA(x0+1, y0, value)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func runDesktopShell(ctx context.Context, requestQuit func()) error {
	return runPlatformDesktopShell(ctx, requestQuit)
}
