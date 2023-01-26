package labeler

import "fmt"

// FormatedLabelScope returns the label in the form of "category/<scope>".
func FormatedLabelScope(scope string) string {
	return fmt.Sprintf("category/%s", scope)
}
