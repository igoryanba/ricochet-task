package workflow

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// RuleEngine движок правил для workflow transitions
type RuleEngine struct {
	conditionEvaluators map[string]ConditionEvaluator
	actionExecutors     map[string]ActionExecutor
	functions           map[string]RuleFunction
	logger              Logger
}

// RuleFunction пользовательская функция для правил
type RuleFunction func(ctx context.Context, args []interface{}) (interface{}, error)

// NewRuleEngine создает новый Rule Engine
func NewRuleEngine(logger Logger) *RuleEngine {
	engine := &RuleEngine{
		conditionEvaluators: make(map[string]ConditionEvaluator),
		actionExecutors:     make(map[string]ActionExecutor),
		functions:           make(map[string]RuleFunction),
		logger:              logger,
	}
	
	// Регистрируем стандартные evaluators
	engine.RegisterConditionEvaluator(&BasicConditionEvaluator{})
	engine.RegisterConditionEvaluator(&TimeConditionEvaluator{})
	engine.RegisterConditionEvaluator(&RegexConditionEvaluator{})
	
	// Регистрируем стандартные executors
	engine.RegisterActionExecutor(&TaskActionExecutor{})
	engine.RegisterActionExecutor(&NotificationActionExecutor{})
	engine.RegisterActionExecutor(&StatusActionExecutor{})
	
	// Регистрируем стандартные функции
	engine.registerStandardFunctions()
	
	return engine
}

// RegisterConditionEvaluator регистрирует оценщик условий
func (re *RuleEngine) RegisterConditionEvaluator(evaluator ConditionEvaluator) {
	// Автоматически определяем тип на основе имени структуры
	evaluatorType := strings.ToLower(strings.TrimSuffix(
		reflect.TypeOf(evaluator).Elem().Name(), "ConditionEvaluator"))
	re.conditionEvaluators[evaluatorType] = evaluator
	re.logger.Info("Registered condition evaluator", "type", evaluatorType)
}

// RegisterActionExecutor регистрирует исполнитель действий
func (re *RuleEngine) RegisterActionExecutor(executor ActionExecutor) {
	re.actionExecutors[executor.GetType()] = executor
	re.logger.Info("Registered action executor", "type", executor.GetType())
}

// RegisterFunction регистрирует пользовательскую функцию
func (re *RuleEngine) RegisterFunction(name string, fn RuleFunction) {
	re.functions[name] = fn
	re.logger.Info("Registered rule function", "name", name)
}

// EvaluateTransitions оценивает переходы workflow на основе события
func (re *RuleEngine) EvaluateTransitions(ctx context.Context, workflow *WorkflowDefinition, event Event) ([]*ActionDefinition, error) {
	var actions []*ActionDefinition
	
	// Создаем контекст для оценки
	evalContext := re.createEvaluationContext(workflow, event)
	
	// Проверяем триггеры
	for _, trigger := range workflow.Triggers {
		if re.matchesTrigger(trigger, event) {
			re.logger.Debug("Trigger matched", "trigger", trigger.Type, "event", event.GetType())
			
			// Оцениваем условия триггера
			if re.evaluateTriggerConditions(ctx, trigger, evalContext) {
				// Добавляем действия триггера
				for _, actionName := range trigger.Actions {
					if action := re.findActionInWorkflow(workflow, actionName); action != nil {
						actions = append(actions, action)
					}
				}
			}
		}
	}
	
	// Проверяем этапы
	for _, stage := range workflow.Stages {
		stageActions := re.evaluateStageTransitions(ctx, stage, evalContext)
		actions = append(actions, stageActions...)
	}
	
	re.logger.Debug("Evaluated transitions", "actionsCount", len(actions))
	return actions, nil
}

// ExecuteAction выполняет действие
func (re *RuleEngine) ExecuteAction(ctx context.Context, action *ActionDefinition, context map[string]interface{}) (map[string]interface{}, error) {
	executor, exists := re.actionExecutors[action.Type]
	if !exists {
		return nil, fmt.Errorf("no executor found for action type: %s", action.Type)
	}
	
	// Проверяем условие действия, если есть
	if action.Condition != "" {
		shouldExecute, err := re.evaluateConditionExpression(ctx, action.Condition, context)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate action condition: %w", err)
		}
		if !shouldExecute {
			re.logger.Debug("Action condition not met, skipping", "action", action.Type)
			return map[string]interface{}{"skipped": true}, nil
		}
	}
	
	re.logger.Debug("Executing action", "type", action.Type)
	return executor.Execute(ctx, action, context)
}

// EvaluateCondition оценивает условие
func (re *RuleEngine) EvaluateCondition(ctx context.Context, condition *ConditionDefinition, context map[string]interface{}) (bool, error) {
	// Определяем тип условия по оператору или полю
	conditionType := re.determineConditionType(condition)
	
	evaluator, exists := re.conditionEvaluators[conditionType]
	if !exists {
		return false, fmt.Errorf("no evaluator found for condition type: %s", conditionType)
	}
	
	return evaluator.Evaluate(ctx, condition, context)
}

// Внутренние методы

func (re *RuleEngine) createEvaluationContext(workflow *WorkflowDefinition, event Event) map[string]interface{} {
	context := map[string]interface{}{
		"event":     event,
		"workflow":  workflow,
		"timestamp": time.Now(),
		"data":      event.GetData(),
	}
	
	// Добавляем переменные workflow
	for k, v := range workflow.Variables {
		context[k] = v
	}
	
	return context
}

func (re *RuleEngine) matchesTrigger(trigger TriggerDefinition, event Event) bool {
	return trigger.Type == event.GetType() || trigger.Type == "*"
}

func (re *RuleEngine) evaluateTriggerConditions(ctx context.Context, trigger TriggerDefinition, context map[string]interface{}) bool {
	for key, value := range trigger.Conditions {
		eventValue := re.getValueFromContext(context, key)
		if !re.compareValues(eventValue, value) {
			return false
		}
	}
	return true
}

func (re *RuleEngine) evaluateStageTransitions(ctx context.Context, stage *StageDefinition, context map[string]interface{}) []*ActionDefinition {
	var actions []*ActionDefinition
	
	// Проверяем условия этапа
	for _, condition := range stage.Conditions {
		result, err := re.EvaluateCondition(ctx, &condition, context)
		if err != nil {
			re.logger.Error("Failed to evaluate stage condition", err, "stage", stage.Name)
			continue
		}
		
		if result {
			// Условие выполнено, добавляем действия этапа
			for i := range stage.Actions {
				actions = append(actions, &stage.Actions[i])
			}
		}
	}
	
	return actions
}

func (re *RuleEngine) findActionInWorkflow(workflow *WorkflowDefinition, actionName string) *ActionDefinition {
	// Ищем действие во всех этапах
	for _, stage := range workflow.Stages {
		for i := range stage.Actions {
			action := &stage.Actions[i]
			if action.Type == actionName {
				return action
			}
		}
	}
	return nil
}

func (re *RuleEngine) determineConditionType(condition *ConditionDefinition) string {
	// Определяем тип по оператору
	switch condition.Operator {
	case "matches", "regex":
		return "regex"
	case "before", "after", "older_than", "newer_than":
		return "time"
	default:
		return "basic"
	}
}

func (re *RuleEngine) evaluateConditionExpression(ctx context.Context, expression string, context map[string]interface{}) (bool, error) {
	// Простой парсер выражений (можно расширить)
	expression = strings.TrimSpace(expression)
	
	// Поддержка функций
	if strings.Contains(expression, "(") {
		return re.evaluateFunction(ctx, expression, context)
	}
	
	// Простое сравнение
	parts := strings.Fields(expression)
	if len(parts) != 3 {
		return false, fmt.Errorf("invalid condition expression: %s", expression)
	}
	
	field := parts[0]
	operator := parts[1]
	value := parts[2]
	
	condition := &ConditionDefinition{
		Field:    field,
		Operator: operator,
		Value:    re.parseValue(value),
	}
	
	return re.EvaluateCondition(ctx, condition, context)
}

func (re *RuleEngine) evaluateFunction(ctx context.Context, expression string, context map[string]interface{}) (bool, error) {
	// Парсим вызов функции
	funcName, args, err := re.parseFunctionCall(expression)
	if err != nil {
		return false, err
	}
	
	fn, exists := re.functions[funcName]
	if !exists {
		return false, fmt.Errorf("unknown function: %s", funcName)
	}
	
	result, err := fn(ctx, args)
	if err != nil {
		return false, err
	}
	
	// Конвертируем результат в boolean
	return re.toBool(result), nil
}

func (re *RuleEngine) parseFunctionCall(expression string) (string, []interface{}, error) {
	// Простой парсер: function_name(arg1, arg2, ...)
	openParen := strings.Index(expression, "(")
	closeParen := strings.LastIndex(expression, ")")
	
	if openParen == -1 || closeParen == -1 || closeParen <= openParen {
		return "", nil, fmt.Errorf("invalid function call syntax: %s", expression)
	}
	
	funcName := strings.TrimSpace(expression[:openParen])
	argsStr := strings.TrimSpace(expression[openParen+1 : closeParen])
	
	var args []interface{}
	if argsStr != "" {
		argParts := strings.Split(argsStr, ",")
		for _, arg := range argParts {
			args = append(args, re.parseValue(strings.TrimSpace(arg)))
		}
	}
	
	return funcName, args, nil
}

func (re *RuleEngine) getValueFromContext(context map[string]interface{}, key string) interface{} {
	// Поддержка вложенных ключей (event.data.field)
	parts := strings.Split(key, ".")
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

func (re *RuleEngine) compareValues(actual, expected interface{}) bool {
	return reflect.DeepEqual(actual, expected)
}

func (re *RuleEngine) parseValue(value string) interface{} {
	// Парсим строковое значение в соответствующий тип
	value = strings.Trim(value, "\"'")
	
	// Boolean
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}
	
	// Number
	if num, err := strconv.Atoi(value); err == nil {
		return num
	}
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		return num
	}
	
	// String
	return value
}

func (re *RuleEngine) toBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != "" && v != "false" && v != "0"
	case int:
		return v != 0
	case float64:
		return v != 0
	default:
		return value != nil
	}
}

func (re *RuleEngine) registerStandardFunctions() {
	// has_tag(tag_name) - проверяет наличие тега
	re.RegisterFunction("has_tag", func(ctx context.Context, args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return false, fmt.Errorf("has_tag requires 1 argument")
		}
		
		tagName, ok := args[0].(string)
		if !ok {
			return false, fmt.Errorf("has_tag argument must be string")
		}
		
		// Получаем теги из контекста (это пример, реальная реализация зависит от структуры данных)
		if event, ok := ctx.Value("event").(Event); ok {
			if data := event.GetData(); data != nil {
				if tags, ok := data["tags"].([]string); ok {
					for _, tag := range tags {
						if tag == tagName {
							return true, nil
						}
					}
				}
			}
		}
		
		return false, nil
	})
	
	// is_business_hours() - проверяет рабочее время
	re.RegisterFunction("is_business_hours", func(ctx context.Context, args []interface{}) (interface{}, error) {
		now := time.Now()
		hour := now.Hour()
		weekday := now.Weekday()
		
		// Рабочие часы: 9-18, пн-пт
		isWeekday := weekday >= time.Monday && weekday <= time.Friday
		isBusinessHour := hour >= 9 && hour < 18
		
		return isWeekday && isBusinessHour, nil
	})
	
	// count_active_tasks() - считает активные задачи
	re.RegisterFunction("count_active_tasks", func(ctx context.Context, args []interface{}) (interface{}, error) {
		// Заглушка - в реальной реализации нужно запросить провайдер
		return 5, nil
	})
}