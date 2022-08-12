package benchmark

import (
	"bytes"
	"log"
	"testing"

	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	zlog "github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/log/format/text"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func BenchmarkVendorLoggers(b *testing.B) {
	const (
		prefix  = "benchmark"
		sub     = "test"
		msg     = "benchmark test log event"
		longMsg = "this is a long message describing a benchmark test log event"
	)

	var (
		meta = map[string]interface{}{
			"complex":  true,
			"id":       1234567890,
			"content":  map[string]interface{}{"data": true},
			"affected": []string{"none", "nothing", "nada"},
		}

		buf             = new(bytes.Buffer)
		logEvent        = event.New().Message(msg).Build()
		logEventComplex = event.New().Prefix(prefix).Sub(sub).Message(longMsg).Metadata(meta).Build()
	)

	b.Run("Writing", func(b *testing.B) {
		b.Run("SimpleText", func(b *testing.B) {
			b.Run("ZeroLogger", func(b *testing.B) {
				localLogger := zerolog.New(buf)

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Info().Msg(msg)
				}
				buf.Reset()
			})
			b.Run("StdLibLogger", func(b *testing.B) {
				localPrefix := prefix + " " // avoid sticking prefix to message
				localLogger := log.New(buf, localPrefix, log.Ltime|log.Lmicroseconds|log.Lmsgprefix)
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Print(msg)
				}
				buf.Reset()
			})
			b.Run("ZapLogger", func(b *testing.B) {
				conf := zap.NewProductionEncoderConfig()
				conf.EncodeTime = zapcore.RFC3339TimeEncoder

				localLogger := zap.New(
					zapcore.NewCore(
						zapcore.NewConsoleEncoder(conf),
						zapcore.AddSync(buf),
						zapcore.DebugLevel,
					),
					zap.AddCaller(),
					zap.WithCaller(false),
				)

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Info(msg)
				}
				buf.Reset()
			})
			b.Run("ZlogLogger", func(b *testing.B) {
				localLogger := zlog.New(
					zlog.WithOut(buf),
					zlog.WithPrefix(prefix),
					zlog.WithFormat(
						text.New().NoLevel().Time(text.LTRFC822Z).Build(),
					),
				)
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Log(logEvent)
				}
				buf.Reset()
			})
			b.Run("LogrusLogger", func(b *testing.B) {
				localLogger := logrus.New()

				localLogger.SetFormatter(&logrus.TextFormatter{
					DisableColors:  true,
					DisableSorting: true,
				})
				localLogger.SetOutput(buf)
				localLogger.SetReportCaller(false)
				localLogger.SetLevel(logrus.InfoLevel)

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Print(msg)
				}
				buf.Reset()
			})
		})
		b.Run("SimpleJSON", func(b *testing.B) {
			b.Run("ZeroLogger", func(b *testing.B) {
				localLogger := zerolog.New(buf)

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Info().Msg(msg)
				}
				buf.Reset()
			})
			b.Run("ZapLogger", func(b *testing.B) {
				conf := zap.NewProductionEncoderConfig()
				conf.EncodeTime = zapcore.RFC3339TimeEncoder

				localLogger := zap.New(
					zapcore.NewCore(
						zapcore.NewJSONEncoder(conf),
						zapcore.AddSync(buf),
						zapcore.DebugLevel,
					),
					zap.AddCaller(),
					zap.WithCaller(false),
				)

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Info(msg)
				}
				buf.Reset()
			})
			b.Run("ZlogLogger", func(b *testing.B) {
				localLogger := zlog.New(
					zlog.WithOut(buf),
					zlog.WithPrefix(prefix),
					zlog.CfgFormatJSONSkipNewline,
				)
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Log(logEvent)
				}
				buf.Reset()
			})
			b.Run("LogrusLogger", func(b *testing.B) {
				localLogger := logrus.New()

				localLogger.SetFormatter(&logrus.JSONFormatter{})
				localLogger.SetOutput(buf)
				localLogger.SetReportCaller(false)
				localLogger.SetLevel(logrus.InfoLevel)

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Print(msg)
				}
				buf.Reset()
			})
		})

		b.Run("ComplexText", func(b *testing.B) {
			b.Run("ZeroLogger", func(b *testing.B) {
				localLogger := zerolog.New(buf)

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Info().Fields(meta).Msg(longMsg)
				}
				buf.Reset()
			})
			b.Run("ZapLogger", func(b *testing.B) {
				conf := zap.NewProductionEncoderConfig()
				conf.EncodeTime = zapcore.RFC3339TimeEncoder

				localLogger := zap.New(
					zapcore.NewCore(
						zapcore.NewConsoleEncoder(conf),
						zapcore.AddSync(buf),
						zapcore.DebugLevel,
					),
					zap.AddCaller(),
					zap.WithCaller(false),
				)

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Info(longMsg, zap.Any("metadata", meta))
				}
				buf.Reset()
			})
			b.Run("ZlogLogger", func(b *testing.B) {
				localLogger := zlog.New(
					zlog.WithOut(buf),
					zlog.WithPrefix(prefix),
					zlog.WithSub(sub),
					zlog.WithFormat(
						text.New().Time(text.LTRFC822Z).Build(),
					),
				)
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Log(logEventComplex)
				}
				buf.Reset()
			})
			b.Run("LogrusLogger", func(b *testing.B) {
				localLogger := logrus.New()

				localLogger.SetFormatter(&logrus.TextFormatter{
					DisableColors:  true,
					DisableSorting: true,
				})
				localLogger.SetOutput(buf)
				localLogger.SetReportCaller(false)
				localLogger.SetLevel(logrus.InfoLevel)

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.WithField("metadata", meta).Print(longMsg)
				}
				buf.Reset()
			})
		})
		b.Run("ComplexJSON", func(b *testing.B) {
			b.Run("ZeroLogger", func(b *testing.B) {
				localLogger := zerolog.New(buf)

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Info().Fields(meta).Msg(longMsg)
				}
				buf.Reset()
			})
			b.Run("ZapLogger", func(b *testing.B) {
				conf := zap.NewProductionEncoderConfig()
				conf.EncodeTime = zapcore.RFC3339TimeEncoder

				localLogger := zap.New(
					zapcore.NewCore(
						zapcore.NewJSONEncoder(conf),
						zapcore.AddSync(buf),
						zapcore.DebugLevel,
					),
					zap.AddCaller(),
					zap.WithCaller(false),
				)

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Info(longMsg, zap.Any("metadata", meta))
				}
				buf.Reset()
			})
			b.Run("ZlogLogger", func(b *testing.B) {
				localLogger := zlog.New(
					zlog.WithOut(buf),
					zlog.WithPrefix(prefix),
					zlog.WithSub(sub),
					zlog.CfgFormatJSONSkipNewline,
				)
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Log(logEventComplex)
				}
				buf.Reset()
			})
			b.Run("LogrusLogger", func(b *testing.B) {
				localLogger := logrus.New()

				localLogger.SetFormatter(&logrus.JSONFormatter{})
				localLogger.SetOutput(buf)
				localLogger.SetReportCaller(false)
				localLogger.SetLevel(logrus.InfoLevel)

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.WithField("metadata", meta).Print(longMsg)
				}
				buf.Reset()
			})
		})
	})

	b.Run("Init", func(b *testing.B) {
		b.Run("SimpleText", func(b *testing.B) {
			b.Run("ZeroLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					l := zerolog.New(buf)
					_ = l
				}
				buf.Reset()
			})
			b.Run("StdLibLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localPrefix := prefix + " " // avoid sticking prefix to message
					l := log.New(buf, localPrefix, log.Ltime|log.Lmicroseconds|log.Lmsgprefix)
					_ = l
				}
				buf.Reset()
			})
			b.Run("ZapLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					conf := zap.NewProductionEncoderConfig()
					conf.EncodeTime = zapcore.RFC3339TimeEncoder

					l := zap.New(
						zapcore.NewCore(
							zapcore.NewConsoleEncoder(conf),
							zapcore.AddSync(buf),
							zapcore.DebugLevel,
						),
						zap.AddCaller(),
						zap.WithCaller(false),
					)
					_ = l
				}
				buf.Reset()
			})
			b.Run("ZlogLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					l := zlog.New(
						zlog.WithOut(buf),
						zlog.WithPrefix(prefix),
						zlog.WithFormat(
							text.New().NoLevel().Time(text.LTRFC822Z).Build(),
						),
					)
					_ = l
				}
				buf.Reset()
			})
			b.Run("LogrusLogger", func(b *testing.B) {

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := logrus.New()

					localLogger.SetFormatter(&logrus.TextFormatter{
						DisableColors:  true,
						DisableSorting: true,
					})
					localLogger.SetOutput(buf)
					localLogger.SetReportCaller(false)
					localLogger.SetLevel(logrus.InfoLevel)
				}
				buf.Reset()
			})
		})
		b.Run("SimpleJSON", func(b *testing.B) {
			b.Run("ZeroLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					l := zerolog.New(buf)
					_ = l
				}
				buf.Reset()
			})
			b.Run("ZapLogger", func(b *testing.B) {

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					conf := zap.NewProductionEncoderConfig()
					conf.EncodeTime = zapcore.RFC3339TimeEncoder

					l := zap.New(
						zapcore.NewCore(
							zapcore.NewJSONEncoder(conf),
							zapcore.AddSync(buf),
							zapcore.DebugLevel,
						),
						zap.AddCaller(),
						zap.WithCaller(false),
					)
					_ = l
				}
				buf.Reset()
			})
			b.Run("ZlogLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					l := zlog.New(
						zlog.WithOut(buf),
						zlog.WithPrefix(prefix),
						zlog.CfgFormatJSONSkipNewline,
					)
					_ = l
				}
				buf.Reset()
			})
			b.Run("LogrusLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := logrus.New()

					localLogger.SetFormatter(&logrus.JSONFormatter{})
					localLogger.SetOutput(buf)
					localLogger.SetReportCaller(false)
					localLogger.SetLevel(logrus.InfoLevel)
				}
				buf.Reset()
			})
		})

		b.Run("ComplexText", func(b *testing.B) {
			b.Run("ZeroLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					l := zerolog.New(buf)
					_ = l
				}
				buf.Reset()
			})
			b.Run("ZapLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					conf := zap.NewProductionEncoderConfig()
					conf.EncodeTime = zapcore.RFC3339TimeEncoder

					l := zap.New(
						zapcore.NewCore(
							zapcore.NewConsoleEncoder(conf),
							zapcore.AddSync(buf),
							zapcore.DebugLevel,
						),
						zap.AddCaller(),
						zap.WithCaller(false),
					)
					_ = l
				}
				buf.Reset()
			})
			b.Run("ZlogLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					l := zlog.New(
						zlog.WithOut(buf),
						zlog.WithPrefix(prefix),
						zlog.WithSub(sub),
						zlog.WithFormat(
							text.New().Time(text.LTRFC822Z).Build(),
						),
					)
					_ = l
				}
				buf.Reset()
			})
			b.Run("LogrusLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := logrus.New()

					localLogger.SetFormatter(&logrus.TextFormatter{
						DisableColors:  true,
						DisableSorting: true,
					})
					localLogger.SetOutput(buf)
					localLogger.SetReportCaller(false)
					localLogger.SetLevel(logrus.InfoLevel)
				}
				buf.Reset()
			})
		})
		b.Run("ComplexJSON", func(b *testing.B) {
			b.Run("ZeroLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					l := zerolog.New(buf)
					_ = l
				}
				buf.Reset()
			})
			b.Run("ZapLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					conf := zap.NewProductionEncoderConfig()
					conf.EncodeTime = zapcore.RFC3339TimeEncoder

					l := zap.New(
						zapcore.NewCore(
							zapcore.NewJSONEncoder(conf),
							zapcore.AddSync(buf),
							zapcore.DebugLevel,
						),
						zap.AddCaller(),
						zap.WithCaller(false),
					)
					_ = l
				}
				buf.Reset()
			})
			b.Run("ZlogLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					l := zlog.New(
						zlog.WithOut(buf),
						zlog.WithPrefix(prefix),
						zlog.WithSub(sub),
						zlog.CfgFormatJSONSkipNewline,
					)
					_ = l
				}
				buf.Reset()
			})
			b.Run("LogrusLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := logrus.New()

					localLogger.SetFormatter(&logrus.JSONFormatter{})
					localLogger.SetOutput(buf)
					localLogger.SetReportCaller(false)
					localLogger.SetLevel(logrus.InfoLevel)
				}
				buf.Reset()
			})
		})
	})

	b.Run("Runtime", func(b *testing.B) {
		b.Run("SimpleText", func(b *testing.B) {
			b.Run("ZeroLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := zerolog.New(buf)
					localLogger.Info().Msg(msg)
				}
				buf.Reset()
			})
			b.Run("StdLibLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localPrefix := prefix + " " // avoid sticking prefix to message
					localLogger := log.New(buf, localPrefix, log.Ltime|log.Lmicroseconds|log.Lmsgprefix)
					localLogger.Print(msg)
				}
				buf.Reset()
			})
			b.Run("ZapLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					conf := zap.NewProductionEncoderConfig()
					conf.EncodeTime = zapcore.RFC3339TimeEncoder

					localLogger := zap.New(
						zapcore.NewCore(
							zapcore.NewConsoleEncoder(conf),
							zapcore.AddSync(buf),
							zapcore.DebugLevel,
						),
						zap.AddCaller(),
						zap.WithCaller(false),
					)
					localLogger.Info(msg)
				}
				buf.Reset()
			})
			b.Run("ZlogLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := zlog.New(
						zlog.WithOut(buf),
						zlog.WithPrefix(prefix),
						zlog.WithFormat(
							text.New().NoLevel().Time(text.LTRFC822Z).Build(),
						),
					)
					localLogger.Log(logEvent)
				}
				buf.Reset()
			})
			b.Run("LogrusLogger", func(b *testing.B) {

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := logrus.New()

					localLogger.SetFormatter(&logrus.TextFormatter{
						DisableColors:  true,
						DisableSorting: true,
					})
					localLogger.SetOutput(buf)
					localLogger.SetReportCaller(false)
					localLogger.SetLevel(logrus.InfoLevel)

					localLogger.Print(msg)
				}
				buf.Reset()
			})
		})
		b.Run("SimpleJSON", func(b *testing.B) {
			b.Run("ZeroLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := zerolog.New(buf)
					localLogger.Info().Msg(msg)
				}
				buf.Reset()
			})
			b.Run("ZapLogger", func(b *testing.B) {

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					conf := zap.NewProductionEncoderConfig()
					conf.EncodeTime = zapcore.RFC3339TimeEncoder

					localLogger := zap.New(
						zapcore.NewCore(
							zapcore.NewJSONEncoder(conf),
							zapcore.AddSync(buf),
							zapcore.DebugLevel,
						),
						zap.AddCaller(),
						zap.WithCaller(false),
					)
					localLogger.Info(msg)
				}
				buf.Reset()
			})
			b.Run("ZlogLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := zlog.New(
						zlog.WithOut(buf),
						zlog.WithPrefix(prefix),
						zlog.CfgFormatJSONSkipNewline,
					)
					localLogger.Log(logEvent)
				}
				buf.Reset()
			})
			b.Run("LogrusLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := logrus.New()

					localLogger.SetFormatter(&logrus.JSONFormatter{})
					localLogger.SetOutput(buf)
					localLogger.SetReportCaller(false)
					localLogger.SetLevel(logrus.InfoLevel)
					localLogger.Print(msg)
				}
				buf.Reset()
			})
		})

		b.Run("ComplexText", func(b *testing.B) {
			b.Run("ZeroLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := zerolog.New(buf)
					localLogger.Info().Fields(meta).Msg(longMsg)
				}
				buf.Reset()
			})
			b.Run("ZapLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					conf := zap.NewProductionEncoderConfig()
					conf.EncodeTime = zapcore.RFC3339TimeEncoder

					localLogger := zap.New(
						zapcore.NewCore(
							zapcore.NewConsoleEncoder(conf),
							zapcore.AddSync(buf),
							zapcore.DebugLevel,
						),
						zap.AddCaller(),
						zap.WithCaller(false),
					)
					localLogger.Info(longMsg, zap.Any("metadata", meta))
				}
				buf.Reset()
			})
			b.Run("ZlogLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := zlog.New(
						zlog.WithOut(buf),
						zlog.WithPrefix(prefix),
						zlog.WithSub(sub),
						zlog.WithFormat(
							text.New().Time(text.LTRFC822Z).Build(),
						),
					)
					localLogger.Log(logEventComplex)
				}
				buf.Reset()
			})
			b.Run("LogrusLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := logrus.New()

					localLogger.SetFormatter(&logrus.TextFormatter{
						DisableColors:  true,
						DisableSorting: true,
					})
					localLogger.SetOutput(buf)
					localLogger.SetReportCaller(false)
					localLogger.SetLevel(logrus.InfoLevel)
					localLogger.WithField("metadata", meta).Print(longMsg)
				}
				buf.Reset()
			})
		})
		b.Run("ComplexJSON", func(b *testing.B) {
			b.Run("ZeroLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := zerolog.New(buf)
					localLogger.Info().Fields(meta).Msg(longMsg)
				}
				buf.Reset()
			})
			b.Run("ZapLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					conf := zap.NewProductionEncoderConfig()
					conf.EncodeTime = zapcore.RFC3339TimeEncoder

					localLogger := zap.New(
						zapcore.NewCore(
							zapcore.NewJSONEncoder(conf),
							zapcore.AddSync(buf),
							zapcore.DebugLevel,
						),
						zap.AddCaller(),
						zap.WithCaller(false),
					)
					localLogger.Info(longMsg, zap.Any("metadata", meta))
				}
				buf.Reset()
			})
			b.Run("ZlogLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := zlog.New(
						zlog.WithOut(buf),
						zlog.WithPrefix(prefix),
						zlog.WithSub(sub),
						zlog.CfgFormatJSONSkipNewline,
					)
					localLogger.Log(logEventComplex)
				}
				buf.Reset()
			})
			b.Run("LogrusLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger := logrus.New()

					localLogger.SetFormatter(&logrus.JSONFormatter{})
					localLogger.SetOutput(buf)
					localLogger.SetReportCaller(false)
					localLogger.SetLevel(logrus.InfoLevel)
					localLogger.WithField("metadata", meta).Print(longMsg)
				}
				buf.Reset()
			})
		})
	})
}
