package studio

import (
	"fmt"

	"github.com/stashapp/stash/pkg/models"
)

type NameExistsError struct {
	Name string
}

func (e *NameExistsError) Error() string {
	return fmt.Sprintf("studio with name '%s' already exists", e.Name)
}

type NameUsedByAliasError struct {
	Name        string
	OtherStudio string
}

func (e *NameUsedByAliasError) Error() string {
	return fmt.Sprintf("name '%s' is used as alias for '%s'", e.Name, e.OtherStudio)
}

// EnsureStudioNameUnique returns an error if the studio name provided
// is used as a name or alias of another existing tag.
func EnsureStudioNameUnique(id int, name string, qb models.StudioReader) error {
	// ensure name is unique
	sameNameStudio, err := ByName(qb, name)
	if err != nil {
		return err
	}

	if sameNameStudio != nil && id != sameNameStudio.ID {
		return &NameExistsError{
			Name: name,
		}
	}

	// query by alias
	sameNameStudio, err = ByAlias(qb, name)
	if err != nil {
		return err
	}

	if sameNameStudio != nil && id != sameNameStudio.ID {
		return &NameUsedByAliasError{
			Name:        name,
			OtherStudio: sameNameStudio.Name.String,
		}
	}

	return nil
}

func EnsureAliasesUnique(id int, aliases []string, qb models.StudioReader) error {
	for _, a := range aliases {
		if err := EnsureStudioNameUnique(id, a, qb); err != nil {
			return err
		}
	}

	return nil
}
