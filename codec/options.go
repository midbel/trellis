package codec

import (
	"fmt"

	"github.com/midbel/sexpr"
	"github.com/midbel/trellis"
)

func makeFlags(opts *OptionsBag) map[string]func() {
	return map[string]func(){
		"border": func() {
			opts.Border = true
		},
		"reverse": func() {
			opts.Reverse = true
		},
		"vertical": func() {
			opts.Orient = trellis.VerticalLayout
		},
		"horizontal": func() {
			opts.Orient = trellis.HorizontalLayout
		},
		"compact": func() {
			opts.Orient = trellis.CompactLayout
		},
	}
}

func makeSetters(opts *OptionsBag) map[string]func(any) error {
	return map[string]func(any) error{
		"width":       assignValue(&opts.Width, parseInt),
		"height":      assignValue(&opts.Height, parseInt),
		"spacing":     assignValue(&opts.Spacing, parseSpacing),
		"border":      assignValue(&opts.Border, parseBool),
		"reverse":     assignValue(&opts.Reverse, parseBool),
		"ticks":       assignValue(&opts.CoordinatesStep, parseInt),
		"connector":   assignValue(&opts.Style, parseConnectorStyle),
		"path":        assignValue(&opts.Path, parsePathType),
		"output":      assignValue(&opts.Type, parseOutput),
		"min-depth":   assignValue(&opts.MinDepth, parseInt),
		"max-depth":   assignValue(&opts.MaxDepth, parseInt),
		"orientation": assignValue(&opts.Orient, parseOrientation),
		"placement":   assignValue(&opts.Placement, parseAlignment),
	}
}

func makeStyleSetters(style *trellis.Style) map[string]func(any) error {
	return map[string]func(any) error{
		"bold":        assignFlag(&style.Flags, trellis.StyleBold),
		"italic":      assignFlag(&style.Flags, trellis.StyleItalic),
		"underline":   assignFlag(&style.Flags, trellis.StyleUnderline),
		"border":      assignValue(&style.Border, parseBool),
		"color":       assignValue(&style.Color, parseString),
		"font-family": assignValue(&style.Family, parseString),
		"font-size":   assignValue(&style.Size, parseNumber),
		"x-align":     assignValue(&style.AlignX, parseAlignment),
		"y-align":     assignValue(&style.AlignY, parseAlignment),
		"margin":      assignValue(&style.Margin, parseInsets),
		"padding":     assignValue(&style.Padding, parseInsets),
	}
}

func assignFlag(dst *trellis.StyleFlag, flag trellis.StyleFlag) func(any) error {
	return func(value any) error {
		v, err := parseBool(value)
		if err != nil {
			return err
		}
		if v {
			*dst |= flag
		} else {
			*dst &^= flag
		}
		return nil
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

func parseSpacing(value any) (int, error) {
	get := func(pad string) (int, error) {
		switch pad {
		case "small":
			return trellis.SpacingS, nil
		case "medium":
			return trellis.SpacingM, nil
		case "large":
			return trellis.SpacingL, nil
		case "extra":
			return trellis.SpacingX, nil
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

func parseInsets(value any) (trellis.Insets, error) {
	get := func(pad string) (int, error) {
		switch pad {
		case "small":
			return trellis.PaddingS, nil
		case "medium":
			return trellis.PaddingM, nil
		case "large":
			return trellis.PaddingL, nil
		default:
			return 0, fmt.Errorf("insets: unsupported value %q", pad)
		}
	}
	var (
		size int
		err  error
	)
	switch v := value.(type) {
	case string:
		size, err = get(v)
	case sexpr.Ident:
		size, err = get(string(v))
	default:
		size, err = parseNumber(value)
	}
	if err != nil {
		return trellis.Insets{}, err
	}
	return trellis.NewInsets(size, size, size, size), nil
}

func parseOrientation(value any) (trellis.Orientation, error) {
	switch v := value.(type) {
	case sexpr.Ident:
		return trellis.ParseOrientation(string(v))
	case string:
		return trellis.ParseOrientation(v)
	default:
		return 0, ErrType
	}
}

func parseOutput(value any) (trellis.Output, error) {
	switch v := value.(type) {
	case sexpr.Ident:
		return trellis.ParseOutput(string(v))
	case string:
		return trellis.ParseOutput(v)
	default:
		return 0, ErrType
	}
}

func parsePathType(value any) (trellis.PathType, error) {
	switch v := value.(type) {
	case sexpr.Ident:
		return trellis.ParsePath(string(v))
	case string:
		return trellis.ParsePath(v)
	default:
		return 0, ErrType
	}
}

func parseConnectorStyle(value any) (trellis.ConnectorStyle, error) {
	switch v := value.(type) {
	case sexpr.Ident:
		return trellis.ParseConnector(string(v))
	case string:
		return trellis.ParseConnector(v)
	default:
		return 0, ErrType
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

func parseNumber(value any) (int, error) {
	if f, ok := value.(float64); ok {
		return int(f), nil
	}
	return parseInt(value)
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
		return false, ErrType
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
