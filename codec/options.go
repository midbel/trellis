package codec

import (
	"fmt"

	"github.com/midbel/sexpr"
	"github.com/midbel/trellis"
)

func makeFlags(opts *trellis.Options) map[string]func() {
	return map[string]func(){
		"border": func() {
			opts.Border = true
		},
		"coordinates": func() {
			opts.ShowCoordinates = true
		},
		"reverse": func() {
			opts.Reverse = true
		},
		"padding": func() {
			opts.Padding++
		},
		"margin": func() {
			opts.Margin++
		},
		"spacing": func() {
			opts.Spacing++
		},
	}
}

func makeSetters(opts *trellis.Options) map[string]func(any) error {
	return map[string]func(any) error{
		"width":       assignValue(&opts.Width, parseInt),
		"height":      assignValue(&opts.Height, parseInt),
		"padding":     assignValue(&opts.Padding, parsePadding),
		"paddingChar": assignValue(&opts.PaddingChar, parseString),
		"margin":      assignValue(&opts.Margin, parseInt),
		"spacing":     assignValue(&opts.Spacing, parseInt),
		"border":      assignValue(&opts.Border, parseBool),
		"reverse":     assignValue(&opts.Reverse, parseBool),
		"coordinates": assignValue(&opts.ShowCoordinates, parseBool),
		"align-x":     assignValue(&opts.AlignX, parseAlignment),
		"align-y":     assignValue(&opts.AlignY, parseAlignment),
	}
}

func assignValue[T any](dst *T, parse func(any) (T, error)) func(any) error {
	return func(value any) error {
		v, err := parse(value)
		if err != nil {
			return err
		}
		*dst = v
		return nil
	}
}

func parsePadding(value any) (int, error) {
	get := func(pad string) (int, error) {
		switch pad {
		case "small":
			return trellis.PaddingS, nil
		case "medium":
			return trellis.PaddingM, nil
		case "large":
			return trellis.PaddingL, nil
		default:
			return 0, fmt.Errorf("padding: unsupported value %q", pad)
		}
	}
	switch v := value.(type) {
	case string:
		return get(v)
	case sexpr.Ident:
		return get(string(v))
	default:
		return parseInt(value)
	}
}

func parseAlignment(value any) (trellis.Alignment, error) {
	switch v := value.(type) {
	case sexpr.Ident:
		return trellis.ParseAlignment(string(v))
	case string:
		return trellis.ParseAlignment(v)
	default:
		return 0, ErrType
	}
}

func parseInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	default:
		return 0, ErrType
	}
}

func parseBool(value any) (bool, error) {
	x, ok := value.(bool)
	if !ok {
		return x, ErrType
	}
	return x, nil
}

func parseString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case sexpr.Ident:
		return string(v), nil
	default:
		return "", ErrType
	}
}
