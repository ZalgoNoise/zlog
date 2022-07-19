package client

import (
	"io"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

type nilLogClient struct{}

// ChanneledLogger impl
func (l *nilLogClient) Close()                                       {}
func (l *nilLogClient) Channels() (chan *event.Event, chan struct{}) { return nil, nil }

// io.Writer impl
func (l *nilLogClient) Write(p []byte) (n int, err error) { return 1, nil }

// log.Logger impl
func (l *nilLogClient) SetOuts(outs ...io.Writer) log.Logger            { return l }
func (l *nilLogClient) AddOuts(outs ...io.Writer) log.Logger            { return l }
func (l *nilLogClient) Prefix(prefix string) log.Logger                 { return l }
func (l *nilLogClient) Sub(sub string) log.Logger                       { return l }
func (l *nilLogClient) Fields(fields map[string]interface{}) log.Logger { return l }
func (l *nilLogClient) IsSkipExit() bool                                { return true }

// log.Printer impl
func (l *nilLogClient) Output(m *event.Event) (n int, err error) { return 1, nil }
func (l *nilLogClient) Log(m ...*event.Event)                    {}
func (l *nilLogClient) Print(v ...interface{})                   {}
func (l *nilLogClient) Println(v ...interface{})                 {}
func (l *nilLogClient) Printf(format string, v ...interface{})   {}
func (l *nilLogClient) Panic(v ...interface{})                   {}
func (l *nilLogClient) Panicln(v ...interface{})                 {}
func (l *nilLogClient) Panicf(format string, v ...interface{})   {}
func (l *nilLogClient) Fatal(v ...interface{})                   {}
func (l *nilLogClient) Fatalln(v ...interface{})                 {}
func (l *nilLogClient) Fatalf(format string, v ...interface{})   {}
func (l *nilLogClient) Error(v ...interface{})                   {}
func (l *nilLogClient) Errorln(v ...interface{})                 {}
func (l *nilLogClient) Errorf(format string, v ...interface{})   {}
func (l *nilLogClient) Warn(v ...interface{})                    {}
func (l *nilLogClient) Warnln(v ...interface{})                  {}
func (l *nilLogClient) Warnf(format string, v ...interface{})    {}
func (l *nilLogClient) Info(v ...interface{})                    {}
func (l *nilLogClient) Infoln(v ...interface{})                  {}
func (l *nilLogClient) Infof(format string, v ...interface{})    {}
func (l *nilLogClient) Debug(v ...interface{})                   {}
func (l *nilLogClient) Debugln(v ...interface{})                 {}
func (l *nilLogClient) Debugf(format string, v ...interface{})   {}
func (l *nilLogClient) Trace(v ...interface{})                   {}
func (l *nilLogClient) Traceln(v ...interface{})                 {}
func (l *nilLogClient) Tracef(format string, v ...interface{})   {}

func NilClient() GRPCLogger {
	return &nilLogClient{}
}
