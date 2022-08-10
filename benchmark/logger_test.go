package log

import (
	"bytes"
	"testing"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/log/format/text"
)

func BenchmarkLogger(b *testing.B) {
	const (
		prefix = "benchmark"
		sub    = "test"
		msg    = "benchmark test log event"
	)

	var (
		meta = map[string]interface{}{
			"complex":  true,
			"id":       1234567890,
			"content":  map[string]interface{}{"data": true},
			"affected": []string{"none", "nothing", "nada"},
		}
		logEventByte    = event.New().Message(msg).Build().Encode()
		logEvent        = event.New().Message(msg).Build()
		logEventComplex = event.New().Prefix(prefix).Sub(sub).Message(msg).Metadata(meta).Build()
		msgByte         = []byte(msg)
		buf             = &bytes.Buffer{}
		reset           = func() {
			buf.Reset()
		}
		logger        = log.New(log.WithOut(buf))
		loggerComplex = log.New(log.WithOut(buf), log.WithPrefix(prefix), log.WithSub(sub), log.WithFormat(log.TextColorLevelFirstSpaced), log.WithFilter(event.Level_trace))
		ml            = log.MultiLogger(
			log.New(log.WithOut(buf)), log.New(log.WithOut(buf)), log.New(log.WithOut(buf)),
			log.New(log.WithOut(buf)), log.New(log.WithOut(buf)), log.New(log.WithOut(buf)),
			log.New(log.WithOut(buf)), log.New(log.WithOut(buf)), log.New(log.WithOut(buf)),
			log.New(log.WithOut(buf)),
		)
		mlc = log.MultiLogger(
			log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)),
			log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)),
			log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)),
			log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)),
		)
	)

	// run benchmarks
	b.Run("Events", func(b *testing.B) {
		b.Run("NewSimpleEvent", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				event.New().Message(msg).Build()
			}
		})
		b.Run("NewSimpleEventWithLevel", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				event.New().Level(event.Level_warn).Message(msg).Build()
			}
		})
		b.Run("NewComplexEvent", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				event.New().Prefix(prefix).Sub(sub).Level(event.Level_warn).Message(msg).Metadata(meta).Build()
			}
		})
		b.Run("NewComplexEventWithCallStack", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				event.New().Prefix(prefix).Sub(sub).Level(event.Level_warn).Message(msg).Metadata(meta).CallStack(true).Build()
			}
		})
	})

	b.Run("Formats", func(b *testing.B) {
		b.Run("TextSimplest", func(b *testing.B) {
			localLogger := log.New(
				log.WithOut(buf),
				log.WithFormat(text.New().NoHeaders().Time(text.LTUnixMilli).Build()),
			)
			for n := 0; n < b.N; n++ {
				localLogger.Log(event.New().Message(msg).Build())
			}
			reset()
		})
		b.Run("TextMostComplex", func(b *testing.B) {
			localLogger := log.New(
				log.WithOut(buf),
				log.WithFormat(text.New().Color().DoubleSpace().Upper().Time(text.LTRFC3339Nano).Build()),
			)
			for n := 0; n < b.N; n++ {
				localLogger.Log(event.New().Message(msg).Build())
			}
			reset()
		})
		b.Run("JSONCompact", func(b *testing.B) {
			localLogger := log.New(
				log.WithOut(buf),
				log.WithFormat(log.FormatJSONSkipNewline),
			)
			for n := 0; n < b.N; n++ {
				localLogger.Log(event.New().Message(msg).Build())
			}
			reset()
		})
		b.Run("JSONIndented", func(b *testing.B) {
			localLogger := log.New(
				log.WithOut(buf),
				log.WithFormat(log.FormatJSONIndent),
			)
			for n := 0; n < b.N; n++ {
				localLogger.Log(event.New().Message(msg).Build())
			}
			reset()
		})
		b.Run("BSON", func(b *testing.B) {
			localLogger := log.New(
				log.WithOut(buf),
				log.WithFormat(log.FormatBSON),
			)
			for n := 0; n < b.N; n++ {
				localLogger.Log(event.New().Message(msg).Build())
			}
			reset()
		})
		b.Run("CSV", func(b *testing.B) {
			localLogger := log.New(
				log.WithOut(buf),
				log.WithFormat(log.FormatCSV),
			)
			for n := 0; n < b.N; n++ {
				localLogger.Log(event.New().Message(msg).Build())
			}
			reset()
		})
		b.Run("XML", func(b *testing.B) {
			localLogger := log.New(
				log.WithOut(buf),
				log.WithFormat(log.FormatXML),
			)
			for n := 0; n < b.N; n++ {
				localLogger.Log(event.New().Message(msg).Build())
			}
			reset()
		})
		b.Run("Gob", func(b *testing.B) {
			localLogger := log.New(
				log.WithOut(buf),
				log.WithFormat(log.FormatGob),
			)
			for n := 0; n < b.N; n++ {
				localLogger.Log(event.New().Message(msg).Build())
			}
			reset()
		})
		b.Run("Protobuf", func(b *testing.B) {
			localLogger := log.New(
				log.WithOut(buf),
				log.WithFormat(log.FormatProtobuf),
			)
			for n := 0; n < b.N; n++ {
				localLogger.Log(event.New().Message(msg).Build())
			}
			reset()
		})
	})

	b.Run("Logger", func(b *testing.B) {
		b.Run("Init", func(b *testing.B) {
			b.Run("NewDefaultLogger", func(b *testing.B) {
				for n := 0; n < b.N; n++ {
					log.New()
				}
			})

			reset()

			b.Run("NewLoggerWithConfig", func(b *testing.B) {
				for n := 0; n < b.N; n++ {
					log.New(
						log.WithPrefix(prefix),
						log.WithSub(sub),
						log.WithFilter(event.Level_warn),
						log.WithFormat(log.TextColor),
						log.WithOut(buf),
					)
				}
			})

			reset()
		})

		b.Run("Writing", func(b *testing.B) {
			b.Run("Write", func(b *testing.B) {
				b.Run("ByteStreamAsInput", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						logger.Write(msgByte)
					}
				})

				reset()

				b.Run("EncodedEventAsInput", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						logger.Write(logEventByte)
					}
				})

				reset()

				b.Run("RawEventAsInput", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						logger.Write(logEvent.Encode())
					}
				})

				reset()
			})

			b.Run("Output", func(b *testing.B) {
				b.Run("SimpleEvent", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						logger.Output(logEvent)
					}
				})

				reset()

				b.Run("ComplexEvent", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						logger.Output(logEventComplex)
					}
				})

				reset()
			})

			b.Run("Print", func(b *testing.B) {
				b.Run("SimpleLogger", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						logger.Print(msg)
					}
				})

				reset()

				b.Run("ComplexLogger", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						loggerComplex.Print(msg)
					}
				})

				reset()
			})
		})
	})

	b.Run("MultiloggerX10", func(b *testing.B) {
		b.Run("Init", func(b *testing.B) {
			b.Run("NewDefaultLogger", func(b *testing.B) {
				for n := 0; n < b.N; n++ {
					log.MultiLogger(
						log.New(), log.New(), log.New(),
						log.New(), log.New(), log.New(),
						log.New(), log.New(), log.New(),
						log.New(),
					)
				}
			})

			reset()

			b.Run("NewLoggerWithConfig", func(b *testing.B) {
				for n := 0; n < b.N; n++ {
					log.MultiLogger(
						log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)),
						log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)),
						log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)),
						log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_warn), log.WithFormat(log.TextColor), log.WithOut(buf)),
					)
				}
			})

			reset()
		})

		b.Run("Writing", func(b *testing.B) {
			b.Run("Write", func(b *testing.B) {
				b.Run("ByteStreamAsInput", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						ml.Write(msgByte)
					}
				})

				reset()

				b.Run("EncodedEventAsInput", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						ml.Write(logEventByte)
					}
				})

				reset()

				b.Run("RawEventAsInput", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						ml.Write(logEvent.Encode())
					}
				})

				reset()

				b.Run("ComplexByteStreamAsInput", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						mlc.Write(msgByte)
					}
				})

				reset()

				b.Run("ComplexEncodedEventAsInput", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						mlc.Write(logEventByte)
					}
				})

				reset()

				b.Run("ComplexRawEventAsInput", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						mlc.Write(logEvent.Encode())
					}
				})

				reset()
			})

			b.Run("Output", func(b *testing.B) {
				b.Run("SimpleEvent", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						ml.Output(logEvent)
					}
				})

				reset()

				b.Run("ComplexEvent", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						ml.Output(logEventComplex)
					}
				})

				reset()

				b.Run("ComplexLoggerSimpleEvent", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						mlc.Output(logEvent)
					}
				})

				reset()

				b.Run("ComplexLoggerComplexEvent", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						mlc.Output(logEventComplex)
					}
				})

				reset()
			})

			b.Run("Print", func(b *testing.B) {
				b.Run("Simple", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						ml.Print(msg)
					}
				})

				reset()

				b.Run("Complex", func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						mlc.Print(msg)
					}
				})

				reset()
			})
		})
	})

	b.Run("Runtime", func(b *testing.B) {
		b.Run("SimpleLoggerPrintCall", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf))
				localLogger.Print(msg)
			}
			reset()
		})
		b.Run("SimpleLoggerLogCall", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf))
				localLogger.Log(event.New().Message(msg).Build())
			}
			reset()
		})
		b.Run("SimpleLoggerWriteString", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf))
				localLogger.Write([]byte(msg))
			}
			reset()
		})
		b.Run("SimpleLoggerWriteEvent", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf))
				localLogger.Write(event.New().Message(msg).Build().Encode())
			}
			reset()
		})
		b.Run("ComplexLoggerPrintCall", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf), log.WithPrefix(prefix), log.WithSub(sub), log.WithFormat(log.TextColorLevelFirstSpaced), log.WithFilter(event.Level_trace))
				localLogger.Print(msg)
			}
			reset()
		})
		b.Run("ComplexLoggerLogCall", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf), log.WithPrefix(prefix), log.WithSub(sub), log.WithFormat(log.TextColorLevelFirstSpaced), log.WithFilter(event.Level_trace))
				localLogger.Log(event.New().Message(msg).Build())
			}
			reset()
		})
		b.Run("ComplexLoggerWriteString", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf), log.WithPrefix(prefix), log.WithSub(sub), log.WithFormat(log.TextColorLevelFirstSpaced), log.WithFilter(event.Level_trace))
				localLogger.Write([]byte(msg))
			}
			reset()
		})
		b.Run("ComplexLoggerWriteEvent", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf), log.WithPrefix(prefix), log.WithSub(sub), log.WithFormat(log.TextColorLevelFirstSpaced), log.WithFilter(event.Level_trace))
				localLogger.Write(event.New().Message(msg).Build().Encode())
			}
			reset()
		})
	})
}
