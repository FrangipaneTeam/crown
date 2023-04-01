package conventionalsizepr

const (
	SizeXS Size = 63 << iota
	SizeS
	SizeM
	SizeL
	SizeXL
)

var Sizes = map[Size]size{
	SizeXS: {
		rangeStart: 0,
		rangeEnd:   49,
		size:       "XS",
	},
	SizeS: {
		rangeStart: 50,
		rangeEnd:   99,
		size:       "S",
	},
	SizeM: {
		rangeStart: 10,
		rangeEnd:   499,
		size:       "M",
	},
	SizeL: {
		rangeStart: 500,
		rangeEnd:   999,
		size:       "L",
	},
	SizeXL: {
		rangeStart: 1000,
		rangeEnd:   100000,
		size:       "XL",
	},
}

type Size int

type size struct {
	rangeStart int
	rangeEnd   int
	size       string
}

// GetRangeStart returns the range start of the size.
func (s Size) GetRangeStart() int {
	return Sizes[s].rangeStart
}

// GetRangeEnd returns the range end of the size.
func (s Size) GetRangeEnd() int {
	return Sizes[s].rangeEnd
}

// IsInRange returns true if the value is in the range.
func (s Size) IsInRange(value int) bool {
	return value >= s.GetRangeStart() && value <= s.GetRangeEnd()
}

// GetSize returns the size name.
// examples : XS, S, M, L, XL.
func (s Size) GetSize() string {
	return Sizes[s].size
}

type PrSize struct {
	addition int
	deletion int
	diff     int
	size     Size
}

// NewPRSize returns a new PRSize.
func NewPRSize(addition, deletion int) *PrSize {
	x := &PrSize{
		addition: addition,
		deletion: deletion,
		diff:     addition + deletion,
	}

	x.defineSize()

	return x
}

// defineSize returns the size of the PR.
func (p *PrSize) defineSize() {
	p.size = SizeXL

	for _, x := range []Size{SizeXS, SizeS, SizeM, SizeL, SizeXL} {
		if x.IsInRange(p.diff) {
			p.size = x
			break
		}
	}
}

// GetAddition returns the addition of the PR.
func (p *PrSize) GetAddition() int {
	return p.addition
}

// GetDeletion returns the deletion of the PR.
func (p *PrSize) GetDeletion() int {
	return p.deletion
}

// GetDiff returns the diff of the PR.
func (p *PrSize) GetDiff() int {
	return p.diff
}

// IsExceeding returns true if the PR is exceeding the limit.
func (p *PrSize) IsTooBig() bool {
	return p.size == SizeXL
}

// GetSize returns the size of the PR.
func (p *PrSize) GetSize() Size {
	return p.size
}
