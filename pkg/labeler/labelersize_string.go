// Code generated by "stringer -type=LabelerSize"; DO NOT EDIT.

package labeler

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[SizeXS-54789]
	_ = x[SizeS-109578]
	_ = x[SizeM-219156]
	_ = x[SizeL-438312]
	_ = x[SizeXL-876624]
}

const (
	_LabelerSize_name_0 = "SizeXS"
	_LabelerSize_name_1 = "SizeS"
	_LabelerSize_name_2 = "SizeM"
	_LabelerSize_name_3 = "SizeL"
	_LabelerSize_name_4 = "SizeXL"
)

func (i LabelerSize) String() string {
	switch {
	case i == 54789:
		return _LabelerSize_name_0
	case i == 109578:
		return _LabelerSize_name_1
	case i == 219156:
		return _LabelerSize_name_2
	case i == 438312:
		return _LabelerSize_name_3
	case i == 876624:
		return _LabelerSize_name_4
	default:
		return "LabelerSize(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
