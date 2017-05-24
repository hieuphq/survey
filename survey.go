package survey

import (
	"errors"

	"github.com/AlecAivazis/survey/core"
)

// Validator is a function passed to a Question in order to redefine
type Validator func(interface{}) error

// Converter is a function passed to a Question in order to convert value types
type Converter func(interface{}) (interface{}, error)

// Question is the core data structure for a survey questionnaire.
type Question struct {
	Name     string
	Prompt   Prompt
	Validate Validator
	Convert  Converter
}

// Prompt is the primary interface for the objects that can take user input
// and return a string value.
type Prompt interface {
	Prompt() (interface{}, error)
	Cleanup(interface{}) error
	Error(error) error
}

// AskOne asks a single question without performing validation on the answer.
func AskOne(p Prompt, t interface{}, v Validator, c Converter) error {
	err := Ask([]*Question{{Prompt: p, Validate: v, Convert: c}}, t)
	if err != nil {
		return err
	}

	return nil
}

// Ask performs the prompt loop
func Ask(qs []*Question, t interface{}) error {

	// if we weren't passed a place to record the answers
	if t == nil {
		// we can't go any further
		return errors.New("cannot call Ask() with a nil reference to record the answers")
	}

	// go over every question
	for _, q := range qs {
		// grab the user input and save it
		ans, err := q.Prompt.Prompt()
		convertedAns := ans
		// if there was a problem
		if err != nil {
			return err
		}

		// if there's a converter
		if q.Convert != nil {
			var invalid error

			// wait for a valid response
			for convertedAns, invalid = q.Convert(ans); invalid != nil; convertedAns, invalid = q.Convert(ans) {
				err := q.Prompt.Error(invalid)
				// if there was a problem
				if err != nil {
					return err
				}

				// ask for more input
				ans, err = q.Prompt.Prompt()
				// if there was a problem
				if err != nil {
					return err
				}
			}
		}

		// if there is a validate handler for this question
		if q.Validate != nil {
			// wait for a valid response
			for invalid := q.Validate(convertedAns); invalid != nil; invalid = q.Validate(convertedAns) {
				err := q.Prompt.Error(invalid)
				// if there was a problem
				if err != nil {
					return err
				}

				// ask for more input
				ans, err = q.Prompt.Prompt()
				// if there was a problem
				if err != nil {
					return err
				}
			}
		}

		// tell the prompt to cleanup with the validated value
		q.Prompt.Cleanup(ans)

		// if something went wrong
		if err != nil {
			// stop listening
			return err
		}

		// add it to the map
		err = core.WriteAnswer(t, q.Name, convertedAns)
		// if something went wrong
		if err != nil {
			return err
		}

	}
	// return the response
	return nil
}
