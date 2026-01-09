package lcd

type Printer interface {
	Print(text string, col, row int) error
}
