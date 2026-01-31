package lcd

type Printer interface {
	Print(text string, col, row int) error
	Rows() int
	Columns() int
	Clear() error
}
