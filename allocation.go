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

func adjustSizes(sizes []Dimension, kind Layout) (Dimension, error) {
	var dim Dimension
	switch kind {
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
		return dim, fmt.Errorf("unsupported layout type")
	}
	return dim, nil
}

func estimateMinSize(root *Node, size int) int {
	count := root.Count()
	if count == 0 {
		return size
	}
	tmp := size / count
	if mod := size % count; mod != 0 {
		tmp++
	}
	return tmp
}

func allocateFromSize(size Dimension, items []*Item, mode Allocate, kind Layout) ([]Dimension, error) {
	if len(items) == 1 {
		return []Dimension{size}, nil
	}
	list := make([]Dimension, 0, len(items))
	switch mode {
	case AllocateEqual:
		if kind == HorizontalLayout {
			height := size.Height / len(items)
			for range items {
				x := NewDimension(size.Width, height)
				list = append(list, x)
			}
		} else if kind == VerticalLayout {
			width := size.Width / len(items)
			for range items {
				x := NewDimension(width, size.Height)
				list = append(list, x)
			}
		} else {
			return nil, fmt.Errorf("unsupported layout type")
		}
	case AllocateProportional:
		var total int
		for i := range items {
			total += items[i].Weight()
		}
		if kind == HorizontalLayout {
			for i := range items {
				h := size.Height * items[i].Weight() / total
				x := NewDimension(size.Width, h)
				list = append(list, x)
			}
		} else if kind == VerticalLayout {
			for i := range items {
				w := size.Width * items[i].Weight() / total
				x := NewDimension(w, size.Height)
				list = append(list, x)
			}
		} else {
			return nil, fmt.Errorf("unsupported layout type")
		}
	default:
		return nil, fmt.Errorf("unsupported allocation mode")
	}
	return list, nil
}
