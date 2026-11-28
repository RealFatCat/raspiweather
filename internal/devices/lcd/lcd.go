package lcd

import (
	"fmt"
	"slices"

	"github.com/realfatcat/lcd1602"
)

var AcceptableFonts = []uint{lcd1602.Font5x8, lcd1602.Font5x10}

type LCD struct {
	lcd  *lcd1602.LCD
	rows int
	cols int
}

func New(bus string, addr int, cols int, rows int, font uint, isBacklightOn bool) (*LCD, error) {
	// Single validation point for the font; callers pass the flag value as-is.
	if !slices.Contains(AcceptableFonts, font) {
		return nil, fmt.Errorf("invalid font value %d, acceptable: %v", font, AcceptableFonts)
	}

	lcdDev, err := lcd1602.New(bus, addr, cols, rows, byte(font), isBacklightOn)
	if err != nil {
		return nil, fmt.Errorf("creating new LCD: %w", err)
	}
	return &LCD{
		lcd:  lcdDev,
		rows: rows,
		cols: cols,
	}, nil
}

func (l *LCD) Close() error {
	return l.lcd.Close()
}

func (l *LCD) Print(text string, row, col int) error {
	return l.lcd.Print(text, row, col)
}

func (l *LCD) Clear() error {
	return l.lcd.Clear()
}

func (l *LCD) ToggleBacklight() error {
	return l.lcd.ToggleBacklight()
}

func (l *LCD) Rows() int {
	return l.rows
}

func (l *LCD) Columns() int {
	return l.cols
}
