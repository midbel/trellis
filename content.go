package trellis

type Content struct {
	Value []rune
	Style
}

func (c Content) String() string {
	return string(c.Value)
}

func (c Content) DisplayWidth() int {
	return DisplayWidth(c.Value)
}

type StyleFlag uint8

const (
	StyleBold StyleFlag = 1 << iota
	StyleItalic
	StyleUnderline
)

type Style struct {
	Flags  StyleFlag
	Size   int
	Family string
	Color  string
}
