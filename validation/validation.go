package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/unstoppablego/framework/logs"
)

// Name of the struct tag used in example.
const tagName = "validate"

// Regular expression to validate email address.
var mailRe = regexp.MustCompile(`\A[\w+\-.]+@[a-z\d\-]+(\.[a-z]+)*\.[a-z]+\z`)

// Generic data validator
type Validator interface {
	//Validate method performs validation and returns results and optional error.
	Validate(interface{}) (bool, error)
}

// DefaultValidator does not perform any validations
type DefaultValidator struct {
}

func (v DefaultValidator) Validate(val interface{}) (bool, error) {
	return true, nil
}

type NumberValidator struct {
	Min int
	Max int
}

func (v NumberValidator) Validate(val interface{}) (bool, error) {
	num := val.(int)

	if num < v.Min {
		return false, fmt.Errorf("should be greater than %v", v.Min)
	}

	if v.Max >= v.Min && num > v.Max {
		return false, fmt.Errorf("should be less than %v", v.Max)
	}

	return true, nil
}

// StringValidator validates string presence and/or its length
type StringValidator struct {
	Min int
	Max int
}

func (v StringValidator) Validate(val interface{}) (bool, error) {
	l := len(val.(string))

	if l == 0 {
		return false, fmt.Errorf("cannot be blank")
	}

	if l < v.Min {
		return false, fmt.Errorf("should be at least %v chars long", v.Min)
	}

	if v.Max >= v.Min && l > v.Max {
		return false, fmt.Errorf("should be less than %v chars long", v.Max)
	}

	return true, nil
}

type EmailValidator struct {
}

func VerifyEmailFormat(email string) bool {
	// pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

func (v EmailValidator) Validate(val interface{}) (bool, error) {
	if !VerifyEmailFormat(val.(string)) {
		return false, fmt.Errorf("is not a valid email address")
	}

	return true, nil
}

type NotEmptyValidator struct {
}

func (v NotEmptyValidator) Validate(val interface{}) (bool, error) {
	if d, ok := val.(string); ok {
		if len(d) <= 0 {
			return false, fmt.Errorf("not be null")
		}
	}
	return true, nil
}

// Returns validator struct corresponding to validation type
func getValidatorFromTag(tag string) []Validator {
	args := strings.Split(tag, ",")
	// log.Println(args)
	var va []Validator

	for i, arg := range args {
		switch arg {
		case "number":
			validator := NumberValidator{}
			fmt.Sscanf(strings.Join(args[1:], ","), "min=%d,max=%d", &validator.Min, &validator.Max)
			va = append(va, validator)
		case "string":
			validator := StringValidator{}
			mm := strings.Split(args[i+1], "-")
			var err error
			validator.Min, err = strconv.Atoi(mm[0])
			if err != nil {
				logs.Error(err)
				break
			}
			validator.Max, err = strconv.Atoi(mm[1])
			if err != nil {
				logs.Error(err)
				break
			}
			// fmt.Sscanf(strings.Join(args[1:], "-"), "min=%d,max=%d", &validator.Min, &validator.Max)
			va = append(va, validator)
		case "email":
			va = append(va, EmailValidator{})
		case "not_empty":
			va = append(va, NotEmptyValidator{})
		}
	}

	return va
}

// Performs actual data validation using validator definitions on the struct
func ValidateStruct(s interface{}) error {
	// errs := []error{}

	// typeofs := reflect.TypeOf(s)
	// logs.Info(typeofs.Name(), typeofs.Kind())
	//ValueOf returns a Value representing the run-time data
	v := reflect.ValueOf(s)

	for i := 0; i < v.NumField(); i++ {
		//Get the field tag value
		tag := v.Type().Field(i).Tag.Get(tagName)

		//Skip if tag is not defined or ignored
		if tag == "" || tag == "-" {
			continue
		}
		//DEBUG
		// fmt.Printf("%d. %v(%v), tag:'%v'\n", i+1, v.Type().Field(i).Name, v.Type().Field(i).Type.Name(), tag)
		//Get a validator that corresponds to a tag
		validator := getValidatorFromTag(tag)

		for _, va := range validator {
			//Perform validation
			valid, err := va.Validate(v.Field(i).Interface())
			//Append error to results
			if !valid && err != nil {
				return fmt.Errorf("%s %s", v.Type().Field(i).Name, err.Error())
				// errs = append(errs, fmt.Errorf("%s %s", v.Type().Field(i).Name, err.Error()))
			}
		}
	}

	return nil
}
