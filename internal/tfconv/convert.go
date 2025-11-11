// Package tfconv provides type conversion utilities for Terraform Plugin Framework.
//
// This package contains helper functions for converting between Go types and
// Terraform Plugin Framework types (attr.Value, types.*). These utilities are
// commonly used when mapping API responses to Terraform resource/data source models.
//
// Key functions:
//   - ConvertMapToObjectValue: Converts map[string]any to attr.Value for types.Dynamic
//   - ConvertToAttrValue: Recursively converts Go values to attr.Value types
//   - Int64Ptr: Converts int to *int64 pointer
package tfconv

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ConvertMapToObjectValue converts a Go map[string]any to an attr.Value for use with types.Dynamic.
// This handles the conversion of arbitrary JSON-like structures to Terraform's type system.
//
// The function normalizes the map through JSON marshaling/unmarshaling to ensure consistent
// type representation, then recursively converts nested structures to appropriate attr.Value types.
//
// Example usage:
//
//	capabilities := map[string]any{"prompts": true, "resources": []string{"file", "http"}}
//	attrVal, err := tfconv.ConvertMapToObjectValue(ctx, capabilities)
//	if err != nil {
//	    return err
//	}
//	data.Capabilities = types.DynamicValue(attrVal)
func ConvertMapToObjectValue(ctx context.Context, m map[string]any) (attr.Value, error) {
	// Convert the map to JSON and back to get a normalized structure
	jsonData, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map; %w", err)
	}

	// Parse into a normalized map
	var normalized map[string]any
	if err := json.Unmarshal(jsonData, &normalized); err != nil {
		return nil, fmt.Errorf("failed to unmarshal map; %w", err)
	}

	// Convert to attr.Value recursively
	return ConvertToAttrValue(normalized)
}

// ConvertToAttrValue recursively converts Go values to attr.Value types.
//
// Supports the following type conversions:
//   - nil -> types.DynamicNull()
//   - string -> types.StringValue()
//   - float64 -> types.NumberValue() (all JSON numbers are float64)
//   - bool -> types.BoolValue()
//   - []any -> types.TupleValue() (recursively converts elements)
//   - map[string]any -> types.ObjectValue() (recursively converts values)
//
// This function is typically used internally by ConvertMapToObjectValue, but can also
// be called directly for converting individual values.
//
// Returns an error if the value type is unsupported or if conversion fails.
func ConvertToAttrValue(v any) (attr.Value, error) {
	switch val := v.(type) {
	case nil:
		return types.DynamicNull(), nil
	case string:
		return types.StringValue(val), nil
	case float64:
		return types.NumberValue(big.NewFloat(val)), nil
	case bool:
		return types.BoolValue(val), nil
	case []any:
		elemTypes := make([]attr.Type, len(val))
		elemValues := make([]attr.Value, len(val))
		for i, elem := range val {
			attrVal, err := ConvertToAttrValue(elem)
			if err != nil {
				return nil, err
			}
			elemTypes[i] = attrVal.Type(context.Background())
			elemValues[i] = attrVal
		}
		tupleVal, diags := types.TupleValue(elemTypes, elemValues)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to create tuple value; %v", diags)
		}
		return tupleVal, nil
	case map[string]any:
		attrTypes := make(map[string]attr.Type)
		attrValues := make(map[string]attr.Value)
		for k, v := range val {
			attrVal, err := ConvertToAttrValue(v)
			if err != nil {
				return nil, err
			}
			attrTypes[k] = attrVal.Type(context.Background())
			attrValues[k] = attrVal
		}
		objVal, diags := types.ObjectValue(attrTypes, attrValues)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to create object value; %v", diags)
		}
		return objVal, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

// Int64Ptr converts an int to *int64 pointer.
//
// This helper is useful when working with API types that use int but Terraform
// expects *int64 (e.g., when using types.Int64PointerValue).
//
// Example usage:
//
//	if gateway.Version != nil {
//	    data.Version = types.Int64PointerValue(tfconv.Int64Ptr(*gateway.Version))
//	}
func Int64Ptr(v int) *int64 {
	i := int64(v)
	return &i
}
