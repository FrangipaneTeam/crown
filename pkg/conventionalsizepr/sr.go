package conventionalsizepr

const (
	SizeXS Size = 63 << iota
	SizeS
	SizeM
	SizeL
	SizeXL
)

var (
	Sizes = map[Size]size{
		SizeXS: {
			count: 10,
			size:  "XS",
		},
		SizeS: {
			count: 50,
			size:  "S",
		},
		SizeM: {
			count: 100,
			size:  "M",
		},
		SizeL: {
			count: 500,
			size:  "L",
		},
		SizeXL: {
			count: 1000,
			size:  "XL",
		},
	}
)

type Size int

type size struct {
	count int
	size  string
}

// GetCount returns the count of the size
func (s Size) GetCount() int {
	return Sizes[s].count
}

// GetSize returns the size name.
// examples : XS, S, M, L, XL
func (s Size) GetSize() string {
	return Sizes[s].size
}

type PrSize struct {
	addition int
	deletion int
	diff     int
	size     Size
}

// NewPRSize returns a new PRSize
func NewPRSize(addition, deletion int) *PrSize {
	x := &PrSize{
		addition: addition,
		deletion: deletion,
		diff:     addition + deletion,
	}

	x.defineSize()

	return x
}

// defineSize returns the size of the PR
func (p *PrSize) defineSize() {

	for _, x := range []Size{SizeXS, SizeS, SizeM, SizeL, SizeXL} {
		if p.diff <= x.GetCount() {
			p.size = x
			return
		}
	}

	p.size = SizeXL
}

// GetAddition returns the addition of the PR
func (p *PrSize) GetAddition() int {
	return p.addition
}

// GetDeletion returns the deletion of the PR
func (p *PrSize) GetDeletion() int {
	return p.deletion
}

// GetDiff returns the diff of the PR
func (p *PrSize) GetDiff() int {
	return p.diff
}

// IsExceeding returns true if the PR is exceeding the limit
func (p *PrSize) IsTooBig() bool {
	return p.size == SizeXL
}

// GetSize returns the size of the PR
func (p *PrSize) GetSize() Size {
	return p.size
}
