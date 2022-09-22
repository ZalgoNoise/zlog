package benchmark

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
		bufs            = []*bytes.Buffer{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}}
		reset           = func() {
			buf.Reset()
			for _, b := range bufs {
				b.Reset()
			}
		}
		logger        = log.New(log.WithOut(buf))
		loggerComplex = log.New(log.WithOut(buf), log.WithPrefix(prefix), log.WithSub(sub), log.WithFormat(log.TextColorLevelFirstSpaced), log.WithFilter(event.Level_trace))
		ml            = log.MultiLogger(
			log.New(log.WithOut(bufs[0])), log.New(log.WithOut(bufs[1])), log.New(log.WithOut(bufs[2])),
			log.New(log.WithOut(bufs[3])), log.New(log.WithOut(bufs[4])), log.New(log.WithOut(bufs[5])),
			log.New(log.WithOut(bufs[6])), log.New(log.WithOut(bufs[7])), log.New(log.WithOut(bufs[8])),
			log.New(log.WithOut(bufs[9])),
		)
		mlc = log.MultiLogger(
			log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[0])), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[1])), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[2])),
			log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[3])), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[4])), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[5])),
			log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[6])), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[7])), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[8])),
			log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[9])),
		)
	)

	// run benchmarks
	b.Run("Events", func(b *testing.B) {
		b.Run("NewSimpleEvent", func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				event.New().Message(msg).Build()
			}
		})
		b.Run("NewSimpleEventWithLevel", func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				event.New().Level(event.Level_warn).Message(msg).Build()
			}
		})
		b.Run("NewComplexEvent", func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				event.New().Prefix(prefix).Sub(sub).Level(event.Level_warn).Message(msg).Metadata(meta).Build()
			}
		})
		b.Run("NewComplexEventWithCallStack", func(b *testing.B) {
			b.ResetTimer()
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
			b.ResetTimer()
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
			b.ResetTimer()
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
			b.ResetTimer()
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
			b.ResetTimer()
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
			b.ResetTimer()
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
			b.ResetTimer()
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
			b.ResetTimer()
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
			b.ResetTimer()
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
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				localLogger.Log(event.New().Message(msg).Build())
			}
			reset()
		})
	})

	b.Run("Logger", func(b *testing.B) {
		b.Run("Init", func(b *testing.B) {
			b.Run("NewDefaultLogger", func(b *testing.B) {
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					log.New()
				}
			})

			reset()

			b.Run("NewLoggerWithConfig", func(b *testing.B) {
				b.ResetTimer()
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
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, err := logger.Write(msgByte)
						if err != nil {
							b.Error(err)
						}
					}
				})

				reset()

				b.Run("EncodedEventAsInput", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, err := logger.Write(logEventByte)
						if err != nil {
							b.Error(err)
						}
					}
				})

				reset()

				b.Run("RawEventAsInput", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, err := logger.Write(logEvent.Encode())
						if err != nil {
							b.Error(err)
						}
					}
				})

				reset()
			})

			b.Run("Output", func(b *testing.B) {
				b.Run("SimpleEvent", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, _ = logger.Output(logEvent)
					}
				})

				reset()

				b.Run("ComplexEvent", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, _ = logger.Output(logEventComplex)
					}
				})

				reset()
			})

			b.Run("Print", func(b *testing.B) {
				b.Run("SimpleLogger", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						logger.Print(msg)
					}
				})

				reset()

				b.Run("ComplexLogger", func(b *testing.B) {
					b.ResetTimer()
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
				b.ResetTimer()
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
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					log.MultiLogger(
						log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[0])), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[1])), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[2])),
						log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[3])), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[4])), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[5])),
						log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[6])), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[7])), log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[8])),
						log.New(log.WithPrefix(prefix), log.WithSub(sub), log.WithFilter(event.Level_trace), log.WithFormat(log.TextColor), log.WithOut(bufs[9])),
					)
				}
			})

			reset()
		})

		b.Run("Writing", func(b *testing.B) {
			b.Run("Write", func(b *testing.B) {
				b.Run("ByteStreamAsInput", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, err := ml.Write(msgByte)
						if err != nil {
							b.Error(err)
						}
					}
				})

				reset()

				b.Run("EncodedEventAsInput", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, err := ml.Write(logEventByte)
						if err != nil {
							b.Error(err)
						}
					}
				})

				reset()

				b.Run("RawEventAsInput", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, err := ml.Write(logEvent.Encode())
						if err != nil {
							b.Error(err)
						}
					}
				})

				reset()

				b.Run("ComplexByteStreamAsInput", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, err := mlc.Write(msgByte)
						if err != nil {
							b.Error(err)
						}
					}
				})

				reset()

				b.Run("ComplexEncodedEventAsInput", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, err := mlc.Write(logEventByte)
						if err != nil {
							b.Error(err)
						}
					}
				})

				reset()

				b.Run("ComplexRawEventAsInput", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, err := mlc.Write(logEvent.Encode())
						if err != nil {
							b.Error(err)
						}
					}
				})

				reset()
			})

			b.Run("Output", func(b *testing.B) {
				b.Run("SimpleEvent", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, err := ml.Output(logEvent)
						if err != nil {
							b.Error(err)
						}
					}
				})

				reset()

				b.Run("ComplexEvent", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, err := ml.Output(logEventComplex)
						if err != nil {
							b.Error(err)
						}
					}
				})

				reset()

				b.Run("ComplexLoggerSimpleEvent", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, err := mlc.Output(logEvent)
						if err != nil {
							b.Error(err)
						}
					}
				})

				reset()

				b.Run("ComplexLoggerComplexEvent", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, err := mlc.Output(logEventComplex)
						if err != nil {
							b.Error(err)
						}
					}
				})

				reset()
			})

			b.Run("Print", func(b *testing.B) {
				b.Run("Simple", func(b *testing.B) {
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						ml.Print(msg)
					}
				})

				reset()

				b.Run("Complex", func(b *testing.B) {
					b.ResetTimer()
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
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf))
				localLogger.Print(msg)
			}
			reset()
		})
		b.Run("SimpleLoggerLogCall", func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf))
				localLogger.Log(event.New().Message(msg).Build())
			}
			reset()
		})
		b.Run("SimpleLoggerWriteString", func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf))
				_, err := localLogger.Write([]byte(msg))
				if err != nil {
					b.Error(err)
				}
			}
			reset()
		})
		b.Run("SimpleLoggerWriteEvent", func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf))
				_, err := localLogger.Write(event.New().Message(msg).Build().Encode())
				if err != nil {
					b.Error(err)
				}
			}
			reset()
		})
		b.Run("ComplexLoggerPrintCall", func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf), log.WithPrefix(prefix), log.WithSub(sub), log.WithFormat(log.TextColorLevelFirstSpaced), log.WithFilter(event.Level_trace))
				localLogger.Print(msg)
			}
			reset()
		})
		b.Run("ComplexLoggerLogCall", func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf), log.WithPrefix(prefix), log.WithSub(sub), log.WithFormat(log.TextColorLevelFirstSpaced), log.WithFilter(event.Level_trace))
				localLogger.Log(event.New().Message(msg).Build())
			}
			reset()
		})
		b.Run("ComplexLoggerWriteString", func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf), log.WithPrefix(prefix), log.WithSub(sub), log.WithFormat(log.TextColorLevelFirstSpaced), log.WithFilter(event.Level_trace))
				_, err := localLogger.Write([]byte(msg))
				if err != nil {
					b.Error(err)
				}
			}
			reset()
		})
		b.Run("ComplexLoggerWriteEvent", func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				localLogger := log.New(log.WithOut(buf), log.WithPrefix(prefix), log.WithSub(sub), log.WithFormat(log.TextColorLevelFirstSpaced), log.WithFilter(event.Level_trace))
				_, err := localLogger.Write(event.New().Message(msg).Build().Encode())
				if err != nil {
					b.Error(err)
				}
			}
			reset()
		})
	})
}
