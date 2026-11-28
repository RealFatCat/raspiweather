package lcd

type Printer interface {
	Print(text string, row, col int) error
	Rows() int
	Columns() int
	Clear() error
}
