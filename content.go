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

func (s Style) Zero() bool {
	return s.Flags == 0 && s.Size == 0 && s.Family == "" && s.Color == ""
}

func (s Style) IsBold() bool {
	return s.Flags&StyleBold != 0
}

func (s Style) IsItalic() bool {
	return s.Flags&StyleItalic != 0
}

func (s Style) IsUnderline() bool {
	return s.Flags&StyleUnderline != 0
}
