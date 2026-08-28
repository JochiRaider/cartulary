package mutationvalue

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

const (
	TargetRecordLink = "record_link"
	TargetRecordTag  = "record_tag"

	OperationCreate = "create"
	OperationPatch  = "patch"
	OperationDelete = "delete"
)

type Value struct {
	targetKind    string
	targetID      string
	operationKind string
	beforeValue   any
	afterValue    any
}

func New(targetKind string, targetID string, operationKind string, beforeValue any, afterValue any) (Value, error) {
	if err := validateTarget(targetKind, targetID); err != nil {
		return Value{}, err
	}
	if err := validateSides(operationKind, beforeValue, afterValue); err != nil {
		return Value{}, err
	}
	var beforeCopy any
	if !isNil(beforeValue) {
		beforeCopy = clone(beforeValue)
	}
	var afterCopy any
	if !isNil(afterValue) {
		afterCopy = clone(afterValue)
	}
	return Value{
		targetKind:    targetKind,
		targetID:      targetID,
		operationKind: operationKind,
		beforeValue:   beforeCopy,
		afterValue:    afterCopy,
	}, nil
}

func (value Value) TargetKind() string {
	return value.targetKind
}

func (value Value) TargetID() string {
	return value.targetID
}

func (value Value) OperationKind() string {
	return value.operationKind
}

func (value Value) BeforeValue() any {
	return clone(value.beforeValue)
}

func (value Value) AfterValue() any {
	return clone(value.afterValue)
}

func Copy(values []Value) []Value {
	result := make([]Value, len(values))
	for index, value := range values {
		result[index], _ = New(value.targetKind, value.targetID, value.operationKind, value.beforeValue, value.afterValue)
	}
	return result
}

func validateTarget(kind string, id string) error {
	switch kind {
	case TargetRecordLink:
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("links: invalid record-link mutation target %q", id)
		}
	case TargetRecordTag:
		parts := strings.Split(id, ":")
		if len(parts) != 3 || parts[0] != TargetRecordTag {
			return fmt.Errorf("links: invalid record-tag mutation target %q", id)
		}
		if _, err := uuid.Parse(parts[1]); err != nil {
			return fmt.Errorf("links: invalid record-tag mutation record target %q", id)
		}
		if _, err := uuid.Parse(parts[2]); err != nil {
			return fmt.Errorf("links: invalid record-tag mutation row target %q", id)
		}
	default:
		return fmt.Errorf("links: invalid mutation target kind %q", kind)
	}
	return nil
}

func validateSides(operation string, beforeValue any, afterValue any) error {
	switch operation {
	case OperationCreate:
		if !isNil(beforeValue) || isNil(afterValue) {
			return fmt.Errorf("links: create mutation requires only an after value")
		}
	case OperationPatch, OperationDelete:
		if isNil(beforeValue) || isNil(afterValue) {
			return fmt.Errorf("links: %s mutation requires before and after values", operation)
		}
	default:
		return fmt.Errorf("links: invalid mutation operation %q", operation)
	}
	return nil
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func clone(value any) any {
	result := cloneReflect(reflect.ValueOf(value))
	if !result.IsValid() {
		return nil
	}
	return result.Interface()
}

func cloneReflect(value reflect.Value) reflect.Value {
	if !value.IsValid() {
		return value
	}
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copy := cloneReflect(value.Elem())
		result := reflect.New(value.Type()).Elem()
		result.Set(copy)
		return result
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.MakeMapWithSize(value.Type(), value.Len())
		iterator := value.MapRange()
		for iterator.Next() {
			result.SetMapIndex(iterator.Key(), cloneReflect(iterator.Value()))
		}
		return result
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for index := 0; index < value.Len(); index++ {
			result.Index(index).Set(cloneReflect(value.Index(index)))
		}
		return result
	case reflect.Array:
		result := reflect.New(value.Type()).Elem()
		for index := 0; index < value.Len(); index++ {
			result.Index(index).Set(cloneReflect(value.Index(index)))
		}
		return result
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.New(value.Type().Elem())
		result.Elem().Set(cloneReflect(value.Elem()))
		return result
	default:
		return value
	}
}
