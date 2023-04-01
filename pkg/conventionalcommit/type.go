package conventionalcommit

const (
	FeatureLabel  CommitType = "feat"
	FixLabel      CommitType = "fix"
	ChoreLabel    CommitType = "chore"
	RefactorLabel CommitType = "refactor"
	DocsLabel     CommitType = "docs"
	StyleLabel    CommitType = "style"
	TestLabel     CommitType = "test"
)

var CommitTypeMap = map[string]CommitType{
	FeatureLabel.String():  FeatureLabel,
	FixLabel.String():      FixLabel,
	ChoreLabel.String():    ChoreLabel,
	RefactorLabel.String(): RefactorLabel,
	DocsLabel.String():     DocsLabel,
	StyleLabel.String():    StyleLabel,
	TestLabel.String():     TestLabel,
}

var CommitTypeList = []string{
	FeatureLabel.String(),
	FixLabel.String(),
	ChoreLabel.String(),
	RefactorLabel.String(),
	DocsLabel.String(),
	StyleLabel.String(),
	TestLabel.String(),
}

func (l *Cc) commitType() {
	switch l.Type() {
	case FeatureLabel.String():
		l.CommitType = FeatureLabel
	case FixLabel.String():
		l.CommitType = FixLabel
	case ChoreLabel.String():
		l.CommitType = ChoreLabel
	case RefactorLabel.String():
		l.CommitType = RefactorLabel
	case DocsLabel.String():
		l.CommitType = DocsLabel
	case StyleLabel.String():
		l.CommitType = StyleLabel
	case TestLabel.String():
		l.CommitType = TestLabel
	default:
		l.CommitType = ""
	}
}

func (l *Cc) IsFeature() bool {
	return l.CommitType == FeatureLabel
}

func (l *Cc) IsFix() bool {
	return l.CommitType == FixLabel
}

func (l *Cc) IsChore() bool {
	return l.CommitType == ChoreLabel
}

func (l *Cc) IsRefactor() bool {
	return l.CommitType == RefactorLabel
}

func (l *Cc) IsDocs() bool {
	return l.CommitType == DocsLabel
}

func (l *Cc) IsStyle() bool {
	return l.CommitType == StyleLabel
}

func (l *Cc) IsTest() bool {
	return l.CommitType == TestLabel
}

func (l *Cc) IsOther() bool {
	return l.CommitType == ""
}
