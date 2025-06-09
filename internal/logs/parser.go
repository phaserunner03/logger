package logs

import (
	"encoding/json"
	"fmt"
	"log"
	"github.com/phaserunner03/logging/configs"
	logpb "google.golang.org/genproto/googleapis/logging/v2"
	logtypepb "google.golang.org/genproto/googleapis/logging/type"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func marshalToJSONString(v interface{}) string {
	if v == nil {
		return "null"
	}

	var data []byte
	var err error

	switch val := v.(type) {
	case *structpb.Struct:
		if val == nil {
			return "null"
		}
		if len(val.GetFields()) == 0 {
			return "{}"
		}
		marshaller := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}
		data, err = marshaller.Marshal(val)
		if err == nil && len(data) == 0 {
			return "{}"
		}

	case *logtypepb.HttpRequest:
		if val == nil {
			return "null"
		}
		marshaller := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}
		data, err = marshaller.Marshal(val)
		if err == nil && len(data) == 0 {
			return "{}"
		}

	case *logpb.LogEntrySourceLocation:
		if val == nil {
			return "null"
		}
		marshaller := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}
		data, err = marshaller.Marshal(val)
		if err == nil && len(data) == 0 {
			return "{}"
		}

	case map[string]string:
		if val == nil {
			return "null"
		}
		if len(val) == 0 {
			return "{}"
		}
		data, err = json.Marshal(val)

	default:
		data, err = json.Marshal(v)
	}

	if err != nil {
		log.Printf("Error marshalling to JSON: %v. Value: %+v. Returning 'null'.", err, v)
		return "null"
	}
	if len(data) == 0 {
		return "null"
	}
	return string(data)
}

func ConvertToBQRow(entry *logpb.LogEntry) (configs.BQLogRow, error) {
	if entry == nil {
		return configs.BQLogRow{}, fmt.Errorf("nil log entry")
	}

	return configs.BQLogRow{
		Timestamp:      entry.GetTimestamp().AsTime(),
		Severity:       entry.GetSeverity().String(),
		LogName:        entry.GetLogName(),
		TextPayload:    entry.GetTextPayload(),
		JsonPayload:    marshalToJSONString(entry.GetJsonPayload()),
		InsertID:       entry.GetInsertId(),
		ResourceType:   entry.GetResource().GetType(),
		ResourceLabels: marshalToJSONString(entry.GetResource().GetLabels()),
		HTTPRequest:    marshalToJSONString(entry.GetHttpRequest()),
		Trace:          entry.GetTrace(),
		SpanID:         entry.GetSpanId(),
		SourceLocation: marshalToJSONString(entry.GetSourceLocation()),
		Labels:         marshalToJSONString(entry.GetLabels()),
	}, nil
}
