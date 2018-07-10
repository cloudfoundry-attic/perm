package logx

type Data struct {
	Key   string
	Value interface{}
}

type Logger interface {
	WithName(name string) Logger
	WithData(data ...Data) Logger

	Debug(msg string, data ...Data)
	Info(msg string, data ...Data)
	Error(msg string, err error, data ...Data)
}
