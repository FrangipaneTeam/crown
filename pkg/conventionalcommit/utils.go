package conventionalcommit

func (l *Cc) IsBreakingChange() bool {
	return l.Commit.IsBreakingChange()
}
