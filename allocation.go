package trellis

import "fmt"

type Allocate uint8

const (
	AllocateEqual Allocate = 1 << iota
	AllocateProportional
)

func ParseAllocate(str string) (Allocate, error) {
	switch str {
	case "equal":
		return AllocateEqual, nil
	case "proportional":
		return AllocateProportional, nil
	default:
		return 0, unknown("allocate", str)
	}
}

func adjustSizes(sizes []Dimension, orient Orientation) (Dimension, error) {
	var dim Dimension
	switch orient {
	case HorizontalLayout:
		for i := range sizes {
			dim.Width = max(sizes[i].Width, dim.Width)
			dim.Height += sizes[i].Height
		}
	case VerticalLayout:
		for i := range sizes {
			dim.Height = max(sizes[i].Height, dim.Height)
			dim.Width += sizes[i].Width
		}
	default:
		return dim, fmt.Errorf("unsupported orientation")
	}
	return dim, nil
}

func estimateMinSize(node *Node, size int) int {
	count := node.Count()
	if count == 0 {
		return size
	}
	tmp := size / count
	if mod := size % count; mod != 0 {
		tmp += mod
	}
	return tmp
}

func allocateFromSize(size Dimension, nodes []*Node, mode Allocate, orient Orientation) ([]Dimension, error) {
	if len(nodes) == 1 {
		return []Dimension{size}, nil
	}
	list := make([]Dimension, 0, len(nodes))
	switch mode {
	case AllocateEqual:
		if orient == HorizontalLayout {
			height := size.Height / len(nodes)
			for range nodes {
				x := NewDimension(size.Width, height)
				list = append(list, x)
			}
		} else if orient == VerticalLayout {
			width := size.Width / len(nodes)
			for range nodes {
				x := NewDimension(width, size.Height)
				list = append(list, x)
			}
		} else {
			return nil, fmt.Errorf("unsupported orientation")
		}
	case AllocateProportional:
		var total int
		for i := range nodes {
			total += nodes[i].Weight()
		}
		if orient == HorizontalLayout {
			for i := range nodes {
				h := size.Height * nodes[i].Weight() / total
				x := NewDimension(size.Width, h)
				list = append(list, x)
			}
		} else if orient == VerticalLayout {
			for i := range nodes {
				w := size.Width * nodes[i].Weight() / total
				x := NewDimension(w, size.Height)
				list = append(list, x)
			}
		} else {
			return nil, fmt.Errorf("unsupported orientation")
		}
	default:
		return nil, fmt.Errorf("unsupported allocation mode")
	}
	return list, nil
}
