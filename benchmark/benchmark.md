### 2022-08-10

```
Running tool: /usr/bin/go test -benchmem -run=^$ -coverprofile=/tmp/vscode-goLfvjOB/go-code-cover -bench . github.com/zalgonoise/zlog/benchmark

goos: linux
goarch: amd64
pkg: github.com/zalgonoise/zlog/benchmark
cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
BenchmarkLogger/Events/NewSimpleEvent-4         	                                  308892	      4572 ns/op	     760 B/op	      18 allocs/op
BenchmarkLogger/Events/NewSimpleEventWithLevel-4         	                          244059	      4539 ns/op	     760 B/op	      18 allocs/op
BenchmarkLogger/Events/NewComplexEvent-4                 	                           28374	     40974 ns/op	    3504 B/op	      98 allocs/op
BenchmarkLogger/Events/NewComplexEventWithCallStack-4    	                            3145	    413149 ns/op	   42751 B/op	     817 allocs/op
BenchmarkLogger/Formats/TextSimplest-4                   	                          168398	      6598 ns/op	    1219 B/op	      27 allocs/op
BenchmarkLogger/Formats/TextMostComplex-4                	                          203071	      6344 ns/op	    1408 B/op	      31 allocs/op
BenchmarkLogger/Formats/JSONCompact-4                    	                          198660	      6718 ns/op	    1135 B/op	      24 allocs/op
BenchmarkLogger/Formats/JSONIndented-4                   	                          124933	      8596 ns/op	    1631 B/op	      26 allocs/op
BenchmarkLogger/Formats/BSON-4                           	                          194053	      5835 ns/op	    1080 B/op	      24 allocs/op
BenchmarkLogger/Formats/CSV-4                            	                          138504	      8685 ns/op	    5120 B/op	      25 allocs/op
BenchmarkLogger/Formats/XML-4                            	                           77788	     16624 ns/op	    5704 B/op	      33 allocs/op
BenchmarkLogger/Formats/Gob-4                            	                           61486	     18881 ns/op	    3440 B/op	      71 allocs/op
BenchmarkLogger/Formats/Protobuf-4                       	                          282685	      4064 ns/op	     876 B/op	      21 allocs/op
BenchmarkLogger/Logger/Init/NewDefaultLogger-4           	                         2176832	     502.7 ns/op	     232 B/op	       4 allocs/op
BenchmarkLogger/Logger/Init/NewLoggerWithConfig-4        	                         1321054	     893.6 ns/op	     364 B/op	       9 allocs/op
BenchmarkLogger/Logger/Writing/Write/ByteStreamAsInput-4 	                          217860	      6347 ns/op	    1376 B/op	      32 allocs/op
BenchmarkLogger/Logger/Writing/Write/EncodedEventAsInput-4         	                  417363	      3281 ns/op	     860 B/op	      19 allocs/op
BenchmarkLogger/Logger/Writing/Write/RawEventAsInput-4             	                  353443	      3620 ns/op	     784 B/op	      20 allocs/op
BenchmarkLogger/Logger/Writing/Output/SimpleEvent-4                	                  670068	      1787 ns/op	     655 B/op	      11 allocs/op
BenchmarkLogger/Logger/Writing/Output/ComplexEvent-4               	                  134132	     10014 ns/op	    3517 B/op	      51 allocs/op
BenchmarkLogger/Logger/Writing/Print/SimpleLogger-4                	                  217400	      5604 ns/op	    1296 B/op	      32 allocs/op
BenchmarkLogger/Logger/Writing/Print/ComplexLogger-4               	                  215510	      5594 ns/op	    1326 B/op	      32 allocs/op
BenchmarkLogger/MultiloggerX10/Init/NewDefaultLogger-4             	                  224998	      4878 ns/op	    2504 B/op	      42 allocs/op
BenchmarkLogger/MultiloggerX10/Init/NewLoggerWithConfig-4          	                  147255	      8655 ns/op	    3824 B/op	      92 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/ByteStreamAsInput-4   	                   20154	     61424 ns/op	   15060 B/op	     320 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/EncodedEventAsInput-4 	                   37996	     31160 ns/op	    8579 B/op	     190 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/RawEventAsInput-4     	                   38583	     31013 ns/op	    7264 B/op	     191 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/ComplexByteStreamAsInput-4         	   20439	     56271 ns/op	   13760 B/op	     310 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/ComplexEncodedEventAsInput-4       	   40593	     30203 ns/op	    7200 B/op	     180 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/ComplexRawEventAsInput-4           	   41760	     31020 ns/op	    7264 B/op	     181 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Output/SimpleEvent-4                     	   77326	     17390 ns/op	    6155 B/op	     110 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Output/ComplexEvent-4                    	   12058	     97516 ns/op	   35170 B/op	     510 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Output/ComplexLoggerSimpleEvent-4        	   69973	     16288 ns/op	    4800 B/op	     100 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Output/ComplexLoggerComplexEvent-4       	   10000	    101558 ns/op	   35013 B/op	     500 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Print/Simple-4                           	   22995	     52611 ns/op	   12816 B/op	     311 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Print/Complex-4                          	   20169	     53758 ns/op	   12816 B/op	     301 allocs/op
BenchmarkLogger/Runtime/SimpleLoggerPrintCall-4                                 	  204382	      5986 ns/op	    1544 B/op	      37 allocs/op
BenchmarkLogger/Runtime/SimpleLoggerLogCall-4                                   	  214236	      6018 ns/op	    1504 B/op	      35 allocs/op
BenchmarkLogger/Runtime/SimpleLoggerWriteString-4                               	  189382	      6355 ns/op	    1648 B/op	      38 allocs/op
BenchmarkLogger/Runtime/SimpleLoggerWriteEvent-4                                	  157904	      7836 ns/op	    1793 B/op	      43 allocs/op
BenchmarkLogger/Runtime/ComplexLoggerPrintCall-4                                	  193396	      6414 ns/op	    1686 B/op	      41 allocs/op
BenchmarkLogger/Runtime/ComplexLoggerLogCall-4                                  	  178197	      6433 ns/op	    1646 B/op	      39 allocs/op
BenchmarkLogger/Runtime/ComplexLoggerWriteString-4                              	  179486	      7154 ns/op	    1790 B/op	      42 allocs/op
BenchmarkLogger/Runtime/ComplexLoggerWriteEvent-4                               	  140169	      8580 ns/op	    1938 B/op	      47 allocs/op
PASS
coverage: [no statements]
ok  	github.com/zalgonoise/zlog/benchmark	69.760s
```
