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

type Style struct {
	Bold      bool
	Italic    bool
	Underline bool
}
