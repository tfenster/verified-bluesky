package shared

import (
	"errors"
)

type ValidationRequest struct {
	BskyHandle string `json:"bskyHandle"`
	VerificationId       string `json:"verificationId"`
}

type TitleAndDescription struct {
	Title       string
	Description string
}

type Naming struct {
	Key                 string
	Title               string
	TitleShortened      string
	Description         string
	FirstAndSecondLevel map[TitleAndDescription][]TitleAndDescription
}

type FlatNaming struct {
	Title	   string
	FirstAndSecondLevel map[string][]string
}

func SetupNamingStructure(moduleSpecifics ModuleSpecifics) (Naming, error) {
	title := "Verified " + moduleSpecifics.ModuleNameShortened
	titleShortened := "Ver. " + moduleSpecifics.ModuleNameShortened
	description := "Verified " + moduleSpecifics.ModuleName
	firstAndSecondLevelTitleAndDesc := map[TitleAndDescription][]TitleAndDescription{}

	for first, secondArray := range moduleSpecifics.FirstAndSecondLevel {
		firstTitle := title + ": " + first
		if len(firstTitle) > 50 {
			firstTitle = titleShortened + ": " + first
		}
		if len(firstTitle) > 50 {
			translated, ok := moduleSpecifics.Level1TranslationMap[first]
			if ok {
				firstTitle = titleShortened + ": " + translated
			}
		}
		if len(firstTitle) > 50 {
			return Naming{}, errors.New("First level title too long: " + firstTitle)
		}
		firstDescription := description + ": " + first
		firstAndSecondLevelTitleAndDesc[TitleAndDescription{Title: firstTitle, Description: firstDescription}] = make([]TitleAndDescription, len(secondArray))

		for i, second := range secondArray {
			secondTitle := title + ": " + first + " - " + second
			secondDescription := description + ": " + first + " - " + second
			if len(secondTitle) > 50 {
				secondTitle = titleShortened + ": " + first + " - " + second
			}
			if len(secondTitle) > 50 {
				translated, ok := moduleSpecifics.Level2TranslationMap[second]
				if ok {
					second = translated
					secondTitle = titleShortened + ": " + first + " - " + second
				}
			}
			if len(secondTitle) > 50 {
				translated, ok := moduleSpecifics.Level1TranslationMap[first]
				if ok {
					secondTitle = titleShortened + ": " + translated + " - " + second
				}
			}
			if len(secondTitle) > 50 {
				return Naming{}, errors.New("Second level title too long: " + secondTitle)
			}
			firstAndSecondLevelTitleAndDesc[TitleAndDescription{Title: firstTitle, Description: firstDescription}][i] = TitleAndDescription{Title: secondTitle, Description: secondDescription}
		}
	}

	return Naming{
		Key:                 moduleSpecifics.ModuleKey,
		Title:               title,
		TitleShortened:      titleShortened,
		Description:         description,
		FirstAndSecondLevel: firstAndSecondLevelTitleAndDesc,
	}, nil
}

func SetupFlatNamingStructure(moduleSpecifics ModuleSpecifics) (FlatNaming, error) {
	naming, err := SetupNamingStructure(moduleSpecifics)
	if err != nil {
		return FlatNaming{}, err
	}
	flatFirstAndSecondLevel := map[string][]string{}
	for first, secondArray := range naming.FirstAndSecondLevel {
		flatFirstAndSecondLevel[first.Title] = make([]string, len(secondArray))
		for i, second := range secondArray {
			flatFirstAndSecondLevel[first.Title][i] = second.Title
		}
	}

	return FlatNaming{
		Title: naming.Title,
		FirstAndSecondLevel: flatFirstAndSecondLevel,
	}, nil
}
