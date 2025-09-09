package workflow

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// BasicConditionEvaluator базовый оценщик условий
type BasicConditionEvaluator struct{}

func (e *BasicConditionEvaluator) Evaluate(ctx context.Context, condition *ConditionDefinition, context map[string]interface{}) (bool, error) {
	fieldValue := e.getFieldValue(condition.Field, context)
	expectedValue := condition.Value
	
	switch condition.Operator {
	case "eq", "equals", "==":
		return e.equals(fieldValue, expectedValue), nil
	case "ne", "not_equals", "!=":
		return !e.equals(fieldValue, expectedValue), nil
	case "gt", ">":
		return e.greaterThan(fieldValue, expectedValue)
	case "gte", ">=":
		return e.greaterThanOrEqual(fieldValue, expectedValue)
	case "lt", "<":
		return e.lessThan(fieldValue, expectedValue)
	case "lte", "<=":
		return e.lessThanOrEqual(fieldValue, expectedValue)
	case "contains":
		return e.contains(fieldValue, expectedValue), nil
	case "in":
		return e.in(fieldValue, expectedValue), nil
	case "starts_with":
		return e.startsWith(fieldValue, expectedValue), nil
	case "ends_with":
		return e.endsWith(fieldValue, expectedValue), nil
	case "is_empty":
		return e.isEmpty(fieldValue), nil
	case "is_not_empty":
		return !e.isEmpty(fieldValue), nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", condition.Operator)
	}
}

func (e *BasicConditionEvaluator) getFieldValue(field string, context map[string]interface{}) interface{} {
	// Поддержка вложенных полей: event.data.status
	parts := strings.Split(field, ".")
	current := context
	
	for i, part := range parts {
		if i == len(parts)-1 {
			return current[part]
		}
		
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}
	
	return nil
}

func (e *BasicConditionEvaluator) equals(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

func (e *BasicConditionEvaluator) greaterThan(a, b interface{}) (bool, error) {
	numA, numB, err := e.convertToNumbers(a, b)
	if err != nil {
		return false, err
	}
	return numA > numB, nil
}

func (e *BasicConditionEvaluator) greaterThanOrEqual(a, b interface{}) (bool, error) {
	numA, numB, err := e.convertToNumbers(a, b)
	if err != nil {
		return false, err
	}
	return numA >= numB, nil
}

func (e *BasicConditionEvaluator) lessThan(a, b interface{}) (bool, error) {
	numA, numB, err := e.convertToNumbers(a, b)
	if err != nil {
		return false, err
	}
	return numA < numB, nil
}

func (e *BasicConditionEvaluator) lessThanOrEqual(a, b interface{}) (bool, error) {
	numA, numB, err := e.convertToNumbers(a, b)
	if err != nil {
		return false, err
	}
	return numA <= numB, nil
}

func (e *BasicConditionEvaluator) contains(a, b interface{}) bool {
	strA := fmt.Sprintf("%v", a)
	strB := fmt.Sprintf("%v", b)
	return strings.Contains(strA, strB)
}

func (e *BasicConditionEvaluator) in(value, list interface{}) bool {
	// Проверяет, находится ли value в списке
	switch l := list.(type) {
	case []interface{}:
		for _, item := range l {
			if reflect.DeepEqual(value, item) {
				return true
			}
		}
	case []string:
		valueStr := fmt.Sprintf("%v", value)
		for _, item := range l {
			if valueStr == item {
				return true
			}
		}
	case string:
		// Список через запятую
		items := strings.Split(l, ",")
		valueStr := fmt.Sprintf("%v", value)
		for _, item := range items {
			if strings.TrimSpace(item) == valueStr {
				return true
			}
		}
	}
	return false
}

func (e *BasicConditionEvaluator) startsWith(a, b interface{}) bool {
	strA := fmt.Sprintf("%v", a)
	strB := fmt.Sprintf("%v", b)
	return strings.HasPrefix(strA, strB)
}

func (e *BasicConditionEvaluator) endsWith(a, b interface{}) bool {
	strA := fmt.Sprintf("%v", a)
	strB := fmt.Sprintf("%v", b)
	return strings.HasSuffix(strA, strB)
}

func (e *BasicConditionEvaluator) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	
	switch v := value.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

func (e *BasicConditionEvaluator) convertToNumbers(a, b interface{}) (float64, float64, error) {
	numA, err := e.toFloat64(a)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot convert %v to number: %w", a, err)
	}
	
	numB, err := e.toFloat64(b)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot convert %v to number: %w", b, err)
	}
	
	return numA, numB, nil
}

func (e *BasicConditionEvaluator) toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("unsupported type: %T", value)
	}
}

// TimeConditionEvaluator оценщик временных условий
type TimeConditionEvaluator struct{}

func (e *TimeConditionEvaluator) Evaluate(ctx context.Context, condition *ConditionDefinition, context map[string]interface{}) (bool, error) {
	fieldValue := e.getFieldValue(condition.Field, context)
	
	timeValue, err := e.parseTime(fieldValue)
	if err != nil {
		return false, fmt.Errorf("failed to parse time from field %s: %w", condition.Field, err)
	}
	
	expectedTime, err := e.parseTime(condition.Value)
	if err != nil {
		return false, fmt.Errorf("failed to parse expected time: %w", err)
	}
	
	switch condition.Operator {
	case "before":
		return timeValue.Before(expectedTime), nil
	case "after":
		return timeValue.After(expectedTime), nil
	case "equals":
		return timeValue.Equal(expectedTime), nil
	case "older_than":
		duration, err := e.parseDuration(condition.Value)
		if err != nil {
			return false, err
		}
		return time.Since(timeValue) > duration, nil
	case "newer_than":
		duration, err := e.parseDuration(condition.Value)
		if err != nil {
			return false, err
		}
		return time.Since(timeValue) < duration, nil
	default:
		return false, fmt.Errorf("unsupported time operator: %s", condition.Operator)
	}
}

func (e *TimeConditionEvaluator) getFieldValue(field string, context map[string]interface{}) interface{} {
	parts := strings.Split(field, ".")
	current := context
	
	for i, part := range parts {
		if i == len(parts)-1 {
			return current[part]
		}
		
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}
	
	return nil
}

func (e *TimeConditionEvaluator) parseTime(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		// Пробуем разные форматы
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		
		return time.Time{}, fmt.Errorf("unable to parse time string: %s", v)
	case int64:
		return time.Unix(v, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported time type: %T", value)
	}
}

func (e *TimeConditionEvaluator) parseDuration(value interface{}) (time.Duration, error) {
	switch v := value.(type) {
	case string:
		return time.ParseDuration(v)
	case int:
		return time.Duration(v) * time.Second, nil
	case int64:
		return time.Duration(v) * time.Second, nil
	default:
		return 0, fmt.Errorf("unsupported duration type: %T", value)
	}
}

// RegexConditionEvaluator оценщик регулярных выражений
type RegexConditionEvaluator struct {
	compiledRegex map[string]*regexp.Regexp
}

func (e *RegexConditionEvaluator) Evaluate(ctx context.Context, condition *ConditionDefinition, context map[string]interface{}) (bool, error) {
	if e.compiledRegex == nil {
		e.compiledRegex = make(map[string]*regexp.Regexp)
	}
	
	fieldValue := e.getFieldValue(condition.Field, context)
	text := fmt.Sprintf("%v", fieldValue)
	
	pattern := fmt.Sprintf("%v", condition.Value)
	
	// Кешируем скомпилированные регексы
	regex, exists := e.compiledRegex[pattern]
	if !exists {
		var err error
		regex, err = regexp.Compile(pattern)
		if err != nil {
			return false, fmt.Errorf("invalid regex pattern %s: %w", pattern, err)
		}
		e.compiledRegex[pattern] = regex
	}
	
	switch condition.Operator {
	case "matches", "regex":
		return regex.MatchString(text), nil
	case "not_matches":
		return !regex.MatchString(text), nil
	default:
		return false, fmt.Errorf("unsupported regex operator: %s", condition.Operator)
	}
}

func (e *RegexConditionEvaluator) getFieldValue(field string, context map[string]interface{}) interface{} {
	parts := strings.Split(field, ".")
	current := context
	
	for i, part := range parts {
		if i == len(parts)-1 {
			return current[part]
		}
		
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}
	
	return nil
}