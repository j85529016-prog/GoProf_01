package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field string
	Err   error
}

// Ошибки обрабатываемых типов.
var (
	ErrInvalidType = errors.New("invalid type") // Тип не является структурой
	ErrEmptyStruct = errors.New("empty struct") // Структура не содержит полей
)

// Ошибки парсинга параметров валидации.
var (
	ErrTagUnsupportedValidator = errors.New("unsupported validator in tag") // Неподдерживаемый валидатор
	ErrTagValidatorEmptyValue  = errors.New("validator value is empty")     // Не указано значение
	ErrEmptyValidator          = errors.New("empty validator in tag")       // Не указан ни один валидатор
)

// Константы валидаторов.
const (
	TagValidate     string = "validate" // Поддерживаемый тег валидации
	NestedValidator string = "nested"   // Валидатор для вложенных структур
)

// Строковые константы поддерживаемых типов.
const (
	TypeStructStr      string = "struct"
	TypeStringStr      string = "string"
	TypeSliceStringStr string = "[]string"
	TypeIntStr         string = "int"
	TypeSliceIntStr    string = "[]int"
)

// Поддерживаемые валидаторы.
var (
	SupportedIntValidators = []string{
		"min",
		"max",
		"in",
	}
	SupportedStringValidators = []string{
		"len",
		"regexp",
		"in",
	}
)

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	lines := make([]string, len(v))
	for i, e := range v {
		lines[i] = fmt.Sprintf("%s: %s", e.Field, e.Err)
	}
	return strings.Join(lines, "\n")
}

//nolint:gocognit
func Validate(v interface{}) error {
	// Ошибки валидации
	var errValidateAllFields ValidationErrors

	structType := reflect.ValueOf(v).Type()
	// Если тип не структура, то ошибка
	if structType.Kind() != reflect.Struct {
		return ErrInvalidType
	}

	// Получаем значение структуры
	structValue := reflect.ValueOf(v)

	structCountFields := structType.NumField()
	// Если структура не содержит полей, то ошибка
	if structCountFields == 0 {
		return ErrEmptyStruct
	}

	// Итерируемся по полям структуры
	for i := 0; i < structCountFields; i++ {
		// Берем поле структуры
		field := structType.Field(i)
		// Если поле не публичное, то пропускаем
		if !field.IsExported() {
			continue
		}

		// Берем имя поля, тип
		fieldName := field.Name
		fieldType := field.Type

		// Проверяем и выводим тип поля
		fieldTypeStr := getSupportedFieldType(fieldType)
		// Если тип не поддерживается, то пропускаем поле
		if fieldTypeStr == "" {
			continue
		}

		// Берем структурный тег поля
		fieldTagValidateValue := field.Tag.Get(TagValidate)
		// Если нет тега валидации, то пропускаем поле
		if fieldTagValidateValue == "" {
			continue
		}

		// Парсим параметры валидации и проверяем их
		validators, err := parseValidateParamsFromTagValue(fieldTagValidateValue, fieldTypeStr)
		if err != nil {
			return err
		}

		// Валидация и добавление ошибки валидации
		fieldValue := structValue.Field(i)
		var errValidateField error
		switch fieldTypeStr {
		case TypeStructStr:
			if fieldTagValidateValue != NestedValidator {
				return ErrTagUnsupportedValidator
			}
			// Рекурсивно вызываем валидацию для вложенных структур
			err = Validate(fieldValue.Interface())
			if err != nil {
				var nestedErrs ValidationErrors
				// Если ошибка валидации то добавляем их сразу во все ошибки валидации
				if errors.As(err, &nestedErrs) {
					for _, e := range nestedErrs {
						errValidateAllFields = append(errValidateAllFields, ValidationError{
							Field: fieldName + "." + e.Field,
							Err:   e.Err,
						})
					}
					continue // Переход к следующему полю
				}
				return err // Программная ошибка
			}
		case TypeIntStr:
			errValidateField = validateInt(int(fieldValue.Int()), validators, fieldName)
		case TypeStringStr:
			errValidateField = validateString(fieldValue.String(), validators, fieldName)
		case TypeSliceIntStr:
			// Явно преобразуем fieldValue к []int для передачи в функцию валидации
			slice := make([]int, fieldValue.Len())
			for j := 0; j < fieldValue.Len(); j++ {
				slice[j] = int(fieldValue.Index(j).Int())
			}
			errValidateField = validateIntSlice(slice, validators, fieldName)
		case TypeSliceStringStr:
			// Явно преобразуем fieldValue к []string для передачи в функцию валидации
			slice := make([]string, fieldValue.Len())
			for j := 0; j < fieldValue.Len(); j++ {
				slice[j] = fieldValue.Index(j).String()
			}
			errValidateField = validateStringSlice(slice, validators, fieldName)
		}

		// Добавляем ошибку валидации поля (если есть) во все ошибки валидации
		if errValidateField != nil {
			errValidateAllFields = append(errValidateAllFields, ValidationError{
				Field: fieldName,
				Err:   errValidateField,
			})
		}
	}

	// Если есть ошибки валидации, то возвращаем их
	if len(errValidateAllFields) > 0 {
		return errValidateAllFields
	}
	return nil
}

func getSupportedFieldType(fieldType reflect.Type) string {
	switch fieldType.Kind() { //nolint:exhaustive
	case reflect.Struct:
		return TypeStructStr
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return TypeIntStr
	case reflect.String:
		return TypeStringStr
	case reflect.Slice:
		switch fieldType.Elem().Kind() { //nolint:exhaustive
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return TypeSliceIntStr
		case reflect.String:
			return TypeSliceStringStr
		default:
			return ""
		}
	default:
		return ""
	}
}

func parseValidateParamsFromTagValue(tagValue string, fieldTypeStr string) (map[string]string, error) {
	// Если валидатор пустой, то ошибка
	if tagValue == "" {
		return nil, ErrEmptyValidator
	}

	// Для struct
	if fieldTypeStr == TypeStructStr {
		if tagValue == NestedValidator {
			return map[string]string{NestedValidator: ""}, nil
		}
		return nil, ErrTagUnsupportedValidator
	}

	tagValuesParts := strings.Split(tagValue, "|")
	mapConstraints := make(map[string]string, len(tagValuesParts))
	for _, part := range tagValuesParts {
		validatorParams := strings.Split(part, ":")

		// Если валидатор не формата name:value, то ошибка
		if len(validatorParams) != 2 {
			return nil, ErrTagUnsupportedValidator
		}

		// Если не поддерживаемый валидатор, то ошибка
		if (fieldTypeStr == TypeStringStr || fieldTypeStr == TypeSliceStringStr) &&
			!slices.Contains(SupportedStringValidators, validatorParams[0]) {
			return nil, ErrTagUnsupportedValidator
		}
		if (fieldTypeStr == TypeIntStr || fieldTypeStr == TypeSliceIntStr) &&
			!slices.Contains(SupportedIntValidators, validatorParams[0]) {
			return nil, ErrTagUnsupportedValidator
		}

		// Если значение валидатора не указано, то ошибка
		if validatorParams[1] == "" {
			return nil, ErrTagValidatorEmptyValue
		}

		// Записываем валидатор и его значение
		mapConstraints[validatorParams[0]] = validatorParams[1]
	}

	return mapConstraints, nil
}

func validateString(value string, rules map[string]string, fieldName string) error {
	if lenVal, ok := rules["len"]; ok {
		expectedLen, err := strconv.Atoi(lenVal)
		if err != nil {
			return fmt.Errorf("invalid len value %q in tag for field %s", lenVal, fieldName)
		}
		if len(value) != expectedLen {
			return fmt.Errorf("length must be %d, got %d", expectedLen, len(value))
		}
	}

	if regexpVal, ok := rules["regexp"]; ok {
		re, err := regexp.Compile(regexpVal)
		if err != nil {
			return fmt.Errorf("invalid regexp value %q in tag for field %s: %w", regexpVal, fieldName, err)
		}
		if !re.MatchString(value) {
			return fmt.Errorf("regexp must match %q, got %q", regexpVal, value)
		}
	}

	if inVal, ok := rules["in"]; ok {
		allowed := strings.Split(inVal, ",")
		if !slices.Contains(allowed, value) {
			return fmt.Errorf("value %q not in allowed list %v", value, allowed)
		}
	}

	return nil
}

func validateStringSlice(values []string, rules map[string]string, fieldName string) error {
	for _, value := range values {
		if err := validateString(value, rules, fieldName); err != nil {
			return fmt.Errorf("element %q in slice does not satisfy validation: %w", value, err)
		}
	}
	return nil
}

func validateInt(value int, rules map[string]string, fieldName string) error {
	if minVal, ok := rules["min"]; ok {
		expectedMin, err := strconv.Atoi(minVal)
		if err != nil {
			return fmt.Errorf("invalid min value %q in tag for field %s: %w", minVal, fieldName, err)
		}
		if value < expectedMin {
			return fmt.Errorf("value %d must be greater than or equal to %d", value, expectedMin)
		}
	}

	if maxVal, ok := rules["max"]; ok {
		expectedMax, err := strconv.Atoi(maxVal)
		if err != nil {
			return fmt.Errorf("invalid max value %q in tag for field %s: %w", maxVal, fieldName, err)
		}
		if value > expectedMax {
			return fmt.Errorf("value %d must be less than or equal to %d", value, expectedMax)
		}
	}

	if inVal, ok := rules["in"]; ok {
		allowedStr := strings.Split(inVal, ",")
		allowedInt := make([]int, len(allowedStr))
		for ix, strVal := range allowedStr {
			intVal, err := strconv.Atoi(strVal)
			if err != nil {
				return fmt.Errorf("non-integer value %q in 'in' constraint for field %s", strVal, fieldName)
			}
			allowedInt[ix] = intVal
		}
		if !slices.Contains(allowedInt, value) {
			return fmt.Errorf("value %d not in allowed set {%s}", value, inVal)
		}
	}
	return nil
}

func validateIntSlice(values []int, rules map[string]string, fieldName string) error {
	for _, value := range values {
		if err := validateInt(value, rules, fieldName); err != nil {
			return fmt.Errorf("element %d in slice does not satisfy validation: %w", value, err)
		}
	}
	return nil
}
