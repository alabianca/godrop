package godrop

const (
	isDirMask  = 2
	isDoneMask = 1
)

type Header struct {
	Size  int64
	Name  string
	Flags int
	Path  string
}

func (h *Header) IsDir() bool {
	if (h.Flags & isDirMask) > 0 {
		return true
	}

	return false
}

func (h *Header) IsComplete() bool {
	if (h.Flags & isDoneMask) > 0 {
		return true
	}

	return false
}

func (h *Header) SetDirBit() {
	h.Flags = h.Flags | isDirMask
}

func (h *Header) SetDoneBit() {
	h.Flags = h.Flags | isDoneMask
}
