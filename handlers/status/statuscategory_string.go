// Code generated by "stringer -type=StatusCategory"; DO NOT EDIT.

package status

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[PR_Check_Title-99]
	_ = x[PR_Check_Commits-198]
	_ = x[PR_Check_SizeChanges-396]
	_ = x[PR_Labeler-792]
	_ = x[Issue_Check_Title-1584]
	_ = x[Issue_Labeler-3168]
}

const (
	_StatusCategory_name_0 = "PR_Check_Title"
	_StatusCategory_name_1 = "PR_Check_Commits"
	_StatusCategory_name_2 = "PR_Check_SizeChanges"
	_StatusCategory_name_3 = "PR_Labeler"
	_StatusCategory_name_4 = "Issue_Check_Title"
	_StatusCategory_name_5 = "Issue_Labeler"
)

func (i StatusCategory) String() string {
	switch {
	case i == 99:
		return _StatusCategory_name_0
	case i == 198:
		return _StatusCategory_name_1
	case i == 396:
		return _StatusCategory_name_2
	case i == 792:
		return _StatusCategory_name_3
	case i == 1584:
		return _StatusCategory_name_4
	case i == 3168:
		return _StatusCategory_name_5
	default:
		return "StatusCategory(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
