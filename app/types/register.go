// app/types/register.go
package types

import (
	"strings"

	"github.com/joeydtaylor/exodus/exodus"
	"github.com/joeydtaylor/exodus/exodus/codec"
	"github.com/joeydtaylor/exodus/exodus/transform"
	"github.com/joeydtaylor/exodus/pkg/electrician"
)

type Feedback struct {
	CustomerID string   `parquet:"name=customerId, type=BYTE_ARRAY, convertedtype=UTF8" json:"customerId"`
	Content    string   `parquet:"name=content, type=BYTE_ARRAY, convertedtype=UTF8" json:"content"`
	Category   string   `parquet:"name=category, type=BYTE_ARRAY, convertedtype=UTF8" json:"category,omitempty"`
	IsNegative bool     `parquet:"name=isNegative, type=BOOLEAN" json:"isNegative"`
	Tags       []string `parquet:"name=tags, type=LIST, valuetype=BYTE_ARRAY, valueconvertedtype=UTF8" json:"tags,omitempty"`
}

func RegisterAll() {
	exodus.MustRegisterType[Feedback]("feedback.v1", codec.JSONStrict)
	electrician.EnableBuilderType[Feedback]("feedback.v1")

	// Manifest-visible transformers for feedback.v1
	transform.Register[Feedback]("feedback.v1", "sentiment", func(f Feedback) (Feedback, error) {
		low := strings.ToLower(f.Content)
		if strings.Contains(low, "love") || strings.Contains(low, "great") || strings.Contains(low, "happy") {
			f.Tags = append(f.Tags, "Positive Sentiment")
		} else {
			f.Tags = append(f.Tags, "Needs Attention")
		}
		return f, nil
	})
	transform.Register[Feedback]("feedback.v1", "tagger", func(f Feedback) (Feedback, error) {
		if f.IsNegative {
			f.Tags = append(f.Tags, "neg")
		}
		return f, nil
	})
	transform.Register[Feedback]("feedback.v1", "audit-only", func(f Feedback) (Feedback, error) {
		// no-op
		return f, nil
	})
}
