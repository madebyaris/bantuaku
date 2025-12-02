package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/bantuaku/backend/errors"
)

// ValidationResult contains the result of validation
type ValidationResult struct {
	Valid  bool
	Errors map[string][]string
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:  true,
		Errors: make(map[string][]string),
	}
}

// AddError adds an error to the validation result
func (vr *ValidationResult) AddError(field, message string) {
	vr.Valid = false
	if vr.Errors[field] == nil {
		vr.Errors[field] = []string{}
	}
	vr.Errors[field] = append(vr.Errors[field], message)
}

// ValidateStruct validates a struct based on validation tags
func ValidateStruct(obj interface{}) *ValidationResult {
	result := NewValidationResult()

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Get validation tags
		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			continue
		}

		// Parse and apply validation rules
		fieldName := fieldType.Name
		if jsonTag := fieldType.Tag.Get("json"); jsonTag != "" {
			fieldName = strings.Split(jsonTag, ",")[0]
		}

		if err := validateField(field.Interface(), tag, result, fieldName); err != nil {
			result.AddError(fieldName, err.Error())
		}
	}

	return result
}

// validateField validates a single field based on validation tags
func validateField(value interface{}, tags string, result *ValidationResult, fieldName string) error {
	if tags == "" {
		return nil
	}

	// Split tags by comma
	ruleTags := strings.Split(tags, ",")

	for _, tag := range ruleTags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}

		// Parse tag and validate
		if err := applyValidationRule(value, tag, fieldName); err != nil {
			return err
		}
	}

	return nil
}

// applyValidationRule applies a single validation rule
func applyValidationRule(value interface{}, tag, fieldName string) error {
	// Parse tag (e.g., "required,email", "min:5", "max:100")
	parts := strings.SplitN(tag, ":", 2)
	rule := parts[0]
	var param string
	if len(parts) > 1 {
		param = parts[1]
	}

	switch rule {
	case "required":
		if isEmpty(value) {
			return fmt.Errorf("%s is required", fieldName)
		}
	case "email":
		if !isEmpty(value) && !isValidEmail(value) {
			return fmt.Errorf("%s must be a valid email address", fieldName)
		}
	case "min":
		if !isEmpty(value) {
			min, err := strconv.Atoi(param)
			if err != nil {
				return fmt.Errorf("invalid min parameter: %s", param)
			}

			if !validateMinLength(value, min) {
				return fmt.Errorf("%s must be at least %d characters", fieldName, min)
			}
		}
	case "max":
		if !isEmpty(value) {
			max, err := strconv.Atoi(param)
			if err != nil {
				return fmt.Errorf("invalid max parameter: %s", param)
			}

			if !validateMaxLength(value, max) {
				return fmt.Errorf("%s must be at most %d characters", fieldName, max)
			}
		}
	case "numeric":
		if !isEmpty(value) && !isNumeric(value) {
			return fmt.Errorf("%s must be numeric", fieldName)
		}
	case "alpha":
		if !isEmpty(value) && !isAlpha(value) {
			return fmt.Errorf("%s must contain only letters", fieldName)
		}
	case "alphanum":
		if !isEmpty(value) && !isAlphaNumeric(value) {
			return fmt.Errorf("%s must contain only letters and numbers", fieldName)
		}
	case "oneof":
		if !isEmpty(value) && !isOneOf(value, strings.Split(param, "|")) {
			return fmt.Errorf("%s must be one of: %s", fieldName, param)
		}
	}

	return nil
}

// isEmpty checks if a value is empty
func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []string:
		return len(v) == 0
	default:
		return reflect.ValueOf(value).IsZero()
	}
}

// isValidEmail checks if a value is a valid email
func isValidEmail(value interface{}) bool {
	s, ok := value.(string)
	if !ok {
		return false
	}

	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(s)
}

// validateMinLength checks if a string value meets minimum length
func validateMinLength(value interface{}, min int) bool {
	s, ok := value.(string)
	if !ok {
		return false
	}

	return utf8.RuneCountInString(s) >= min
}

// validateMaxLength checks if a string value meets maximum length
func validateMaxLength(value interface{}, max int) bool {
	s, ok := value.(string)
	if !ok {
		return false
	}

	return utf8.RuneCountInString(s) <= max
}

// isNumeric checks if a value is numeric
func isNumeric(value interface{}) bool {
	s, ok := value.(string)
	if !ok {
		return false
	}

	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// isAlpha checks if a value contains only letters
func isAlpha(value interface{}) bool {
	s, ok := value.(string)
	if !ok {
		return false
	}

	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	return alphaRegex.MatchString(s)
}

// isAlphaNumeric checks if a value contains only letters and numbers
func isAlphaNumeric(value interface{}) bool {
	s, ok := value.(string)
	if !ok {
		return false
	}

	alphanumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	return alphanumRegex.MatchString(s)
}

// isOneOf checks if a value is one of the allowed values
func isOneOf(value interface{}, allowed []string) bool {
	s, ok := value.(string)
	if !ok {
		return false
	}

	for _, allowedValue := range allowed {
		if s == allowedValue {
			return true
		}
	}

	return false
}

// Validate validates a struct and returns an error if validation fails
func Validate(obj interface{}) error {
	result := ValidateStruct(obj)

	if !result.Valid {
		// Convert validation errors to a single error
		var errorMessages []string
		for field, messages := range result.Errors {
			for _, message := range messages {
				errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", field, message))
			}
		}

		return errors.NewValidationError("Validation failed", strings.Join(errorMessages, "; "))
	}

	return nil
}

// ValidateWithResult validates a struct and returns the detailed result
func ValidateWithResult(obj interface{}) *ValidationResult {
	return ValidateStruct(obj)
}
