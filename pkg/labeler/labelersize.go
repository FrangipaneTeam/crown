package labeler

import (
	"fmt"

	"github.com/google/go-github/v47/github"

	"github.com/FrangipaneTeam/crown/pkg/conventionalsizepr"
)

const (
	// ! Always add new IDs at the END of the list.
	SizeXS LabelerSize = 54789 << iota
	SizeS
	SizeM
	SizeL
	SizeXL
	// ! Always add new IDs at the END of the list.
)

const (
	prefixSize = "size"
)

//go:generate stringer -type=LabelerSize
type LabelerSize int //nolint:revive

type size struct {
	code     conventionalsizepr.Size
	longName string
	color    string
}

var LabelsSize = map[LabelerSize]size{
	SizeXS: {
		code:     conventionalsizepr.SizeXS,
		longName: fmt.Sprintf("%s/%s", prefixSize, conventionalsizepr.SizeXS.GetSize()),
		color:    "2cbe4e",
	},
	SizeS: {
		code:     conventionalsizepr.SizeS,
		longName: fmt.Sprintf("%s/%s", prefixSize, conventionalsizepr.SizeS.GetSize()),
		color:    "2cbe4e",
	},
	SizeM: {
		code:     conventionalsizepr.SizeM,
		longName: fmt.Sprintf("%s/%s", prefixSize, conventionalsizepr.SizeM.GetSize()),
		color:    "fe7d37",
	},
	SizeL: {
		code:     conventionalsizepr.SizeL,
		longName: fmt.Sprintf("%s/%s", prefixSize, conventionalsizepr.SizeL.GetSize()),
		color:    "e05d44",
	},
	SizeXL: {
		code:     conventionalsizepr.SizeXL,
		longName: fmt.Sprintf("%s/%s", prefixSize, conventionalsizepr.SizeXL.GetSize()),
		color:    "ff0000",
	},
}

// FindLabelerSize returns the labeler size from a conventional size.
func FindLabelerSize(size conventionalsizepr.Size) LabelerSize {
	for k, v := range LabelsSize {
		if v.code == size {
			return k
		}
	}
	return -1
}

// GetLongName returns the long name of the label.
func (c LabelerSize) GetLongName() string {
	return LabelsSize[c].longName
}

// GetCode returns the code of the label.
func (c LabelerSize) GetCode() conventionalsizepr.Size {
	return LabelsSize[c].code
}

// GithubLabel returns the github label of the label.
func (c LabelerSize) GithubLabel() github.Label {
	return github.Label{
		Name:  github.String(LabelsSize[c].longName),
		Color: github.String(LabelsSize[c].color),
	}
}
