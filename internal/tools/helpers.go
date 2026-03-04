package tools

import (
	"time"

	date "google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func argStr(args map[string]any, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

func argInt64(args map[string]any, key string) *int64 {
	if v, ok := args[key].(float64); ok {
		i := int64(v)
		return &i
	}
	return nil
}

func argInt32(args map[string]any, key string) *int32 {
	if v, ok := args[key].(float64); ok {
		i := int32(v)
		return &i
	}
	return nil
}

func argBool(args map[string]any, key string) *bool {
	if v, ok := args[key].(bool); ok {
		return &v
	}
	return nil
}

func argDate(args map[string]any, key string) *date.Date {
	s := argStr(args, key)
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &date.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}

func argTimestamp(args map[string]any, key string) *timestamppb.Timestamp {
	s := argStr(args, key)
	if s == "" {
		return nil
	}
	if len(s) == 10 {
		s += "T00:00:00Z"
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return timestamppb.New(t)
}
