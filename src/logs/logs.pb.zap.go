// Code generated by protoc-gen-zap (etc/proto/protoc-gen-zap). DO NOT EDIT.
//
// source: logs/logs.proto

package logs

import (
	fmt "fmt"
	protoextensions "github.com/pachyderm/pachyderm/v2/src/protoextensions"
	zapcore "go.uber.org/zap/zapcore"
)

func (x *LogQuery) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	if obj, ok := interface{}(x.GetUser()).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("user", obj)
	} else {
		enc.AddReflected("user", x.GetUser())
	}
	if obj, ok := interface{}(x.GetAdmin()).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("admin", obj)
	} else {
		enc.AddReflected("admin", x.GetAdmin())
	}
	return nil
}

func (x *AdminLogQuery) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("logql", x.GetLogql())
	enc.AddString("pod", x.GetPod())
	if obj, ok := interface{}(x.GetPodContainer()).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("pod_container", obj)
	} else {
		enc.AddReflected("pod_container", x.GetPodContainer())
	}
	enc.AddString("app", x.GetApp())
	if obj, ok := interface{}(x.GetMaster()).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("master", obj)
	} else {
		enc.AddReflected("master", x.GetMaster())
	}
	if obj, ok := interface{}(x.GetStorage()).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("storage", obj)
	} else {
		enc.AddReflected("storage", x.GetStorage())
	}
	if obj, ok := interface{}(x.GetUser()).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("user", obj)
	} else {
		enc.AddReflected("user", x.GetUser())
	}
	return nil
}

func (x *PodContainer) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("pod", x.Pod)
	enc.AddString("container", x.Container)
	return nil
}

func (x *UserLogQuery) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("project", x.GetProject())
	if obj, ok := interface{}(x.GetPipeline()).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("pipeline", obj)
	} else {
		enc.AddReflected("pipeline", x.GetPipeline())
	}
	enc.AddString("datum", x.GetDatum())
	enc.AddString("job", x.GetJob())
	if obj, ok := interface{}(x.GetPipelineJob()).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("pipeline_job", obj)
	} else {
		enc.AddReflected("pipeline_job", x.GetPipelineJob())
	}
	return nil
}

func (x *PipelineLogQuery) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("project", x.Project)
	enc.AddString("pipeline", x.Pipeline)
	return nil
}

func (x *PipelineJobLogQuery) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	if obj, ok := interface{}(x.Pipeline).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("pipeline", obj)
	} else {
		enc.AddReflected("pipeline", x.Pipeline)
	}
	enc.AddString("job", x.Job)
	return nil
}

func (x *PipelineDatumLogQuery) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	if obj, ok := interface{}(x.Pipeline).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("pipeline", obj)
	} else {
		enc.AddReflected("pipeline", x.Pipeline)
	}
	enc.AddString("datum", x.Datum)
	return nil
}

func (x *LogFilter) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	if obj, ok := interface{}(x.TimeRange).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("time_range", obj)
	} else {
		enc.AddReflected("time_range", x.TimeRange)
	}
	enc.AddUint64("limit", x.Limit)
	if obj, ok := interface{}(x.Regex).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("regex", obj)
	} else {
		enc.AddReflected("regex", x.Regex)
	}
	enc.AddString("level", x.Level.String())
	return nil
}

func (x *TimeRangeLogFilter) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	protoextensions.AddTimestamp(enc, "from", x.From)
	protoextensions.AddTimestamp(enc, "until", x.Until)
	return nil
}

func (x *RegexLogFilter) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("pattern", x.Pattern)
	enc.AddBool("negate", x.Negate)
	return nil
}

func (x *GetLogsRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	if obj, ok := interface{}(x.Query).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("query", obj)
	} else {
		enc.AddReflected("query", x.Query)
	}
	if obj, ok := interface{}(x.Filter).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("filter", obj)
	} else {
		enc.AddReflected("filter", x.Filter)
	}
	enc.AddBool("tail", x.Tail)
	enc.AddBool("want_paging_hint", x.WantPagingHint)
	enc.AddString("log_format", x.LogFormat.String())
	return nil
}

func (x *GetLogsResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	if obj, ok := interface{}(x.GetPagingHint()).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("paging_hint", obj)
	} else {
		enc.AddReflected("paging_hint", x.GetPagingHint())
	}
	if obj, ok := interface{}(x.GetLog()).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("log", obj)
	} else {
		enc.AddReflected("log", x.GetLog())
	}
	return nil
}

func (x *PagingHint) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	if obj, ok := interface{}(x.Older).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("older", obj)
	} else {
		enc.AddReflected("older", x.Older)
	}
	if obj, ok := interface{}(x.Newer).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("newer", obj)
	} else {
		enc.AddReflected("newer", x.Newer)
	}
	return nil
}

func (x *LogMessage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	if obj, ok := interface{}(x.GetVerbatim()).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("verbatim", obj)
	} else {
		enc.AddReflected("verbatim", x.GetVerbatim())
	}
	if obj, ok := interface{}(x.GetJson()).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("json", obj)
	} else {
		enc.AddReflected("json", x.GetJson())
	}
	enc.AddObject("pps_log_message", x.GetPpsLogMessage())
	return nil
}

func (x *VerbatimLogMessage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	protoextensions.AddBytes(enc, "line", x.Line)
	protoextensions.AddTimestamp(enc, "timestamp", x.Timestamp)
	return nil
}

func (x *ParsedJSONLogMessage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	if obj, ok := interface{}(x.Verbatim).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("verbatim", obj)
	} else {
		enc.AddReflected("verbatim", x.Verbatim)
	}
	enc.AddObject("fields", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		for k, v := range x.Fields {
			enc.AddString(fmt.Sprintf("%v", k), v)
		}
		return nil
	}))
	protoextensions.AddTimestamp(enc, "native_timestamp", x.NativeTimestamp)
	enc.AddObject("pps_log_message", x.PpsLogMessage)
	return nil
}
