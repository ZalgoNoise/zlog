### 2022-08-12 - Intel i5-4300M CPU @ 2.60GHz

#### [`logger_test.go`](./logger_test.go)

```
Running tool: /usr/bin/go test -benchmem -run=^$ -coverprofile=/tmp/vscode-goqXOu1q/go-code-cover -bench ^BenchmarkLogger$ github.com/zalgonoise/zlog/benchmark

goos: linux
goarch: amd64
pkg: github.com/zalgonoise/zlog/benchmark
cpu: Intel(R) Core(TM) i5-4300M CPU @ 2.60GHz
BenchmarkLogger/Events/NewSimpleEvent-4         	                                  726506	      3327 ns/op	     760 B/op	      18 allocs/op
BenchmarkLogger/Events/NewSimpleEventWithLevel-4         	                          560275	      2211 ns/op	     760 B/op	      18 allocs/op
BenchmarkLogger/Events/NewComplexEvent-4                 	                           68781	     20051 ns/op	    3248 B/op	      96 allocs/op
BenchmarkLogger/Events/NewComplexEventWithCallStack-4    	                            3981	    273390 ns/op	   40843 B/op	     803 allocs/op
BenchmarkLogger/Formats/TextSimplest-4                   	                          410092	      2704 ns/op	    1137 B/op	      27 allocs/op
BenchmarkLogger/Formats/TextMostComplex-4                	                          391533	      3127 ns/op	    1264 B/op	      31 allocs/op
BenchmarkLogger/Formats/JSONCompact-4                    	                          335636	      3752 ns/op	    1335 B/op	      24 allocs/op
BenchmarkLogger/Formats/JSONIndented-4                   	                          251832	      4836 ns/op	    1583 B/op	      27 allocs/op
BenchmarkLogger/Formats/BSON-4                           	                          367041	      3150 ns/op	    1080 B/op	      24 allocs/op
BenchmarkLogger/Formats/CSV-4                            	                          300316	      3743 ns/op	    5120 B/op	      25 allocs/op
BenchmarkLogger/Formats/XML-4                            	                          130011	      7959 ns/op	    5704 B/op	      33 allocs/op
BenchmarkLogger/Formats/Gob-4                            	                          126925	      8926 ns/op	    3200 B/op	      70 allocs/op
BenchmarkLogger/Formats/Protobuf-4                       	                          566678	      2122 ns/op	     874 B/op	      21 allocs/op
BenchmarkLogger/Logger/Init/NewDefaultLogger-4           	                         4415923	     263.8 ns/op	     232 B/op	       4 allocs/op
BenchmarkLogger/Logger/Init/NewLoggerWithConfig-4        	                         2194132	     487.7 ns/op	     364 B/op	       9 allocs/op
BenchmarkLogger/Logger/Writing/Write/ByteStreamAsInput-4 	                          405830	      3059 ns/op	    1376 B/op	      32 allocs/op
BenchmarkLogger/Logger/Writing/Write/EncodedEventAsInput-4         	                  648042	      2242 ns/op	     720 B/op	      19 allocs/op
BenchmarkLogger/Logger/Writing/Write/RawEventAsInput-4             	                  548157	      2219 ns/op	     768 B/op	      20 allocs/op
BenchmarkLogger/Logger/Writing/Output/SimpleEvent-4                	                 1000000	      1009 ns/op	     614 B/op	      11 allocs/op
BenchmarkLogger/Logger/Writing/Output/ComplexEvent-4               	                  209353	      6224 ns/op	    3517 B/op	      51 allocs/op
BenchmarkLogger/Logger/Writing/Print/SimpleLogger-4                	                  432414	      2902 ns/op	    1296 B/op	      32 allocs/op
BenchmarkLogger/Logger/Writing/Print/ComplexLogger-4               	                  399728	      2898 ns/op	    1326 B/op	      32 allocs/op
BenchmarkLogger/MultiloggerX10/Init/NewDefaultLogger-4             	                  428790	      2675 ns/op	    2504 B/op	      42 allocs/op
BenchmarkLogger/MultiloggerX10/Init/NewLoggerWithConfig-4          	                  206586	      5062 ns/op	    3824 B/op	      92 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/ByteStreamAsInput-4   	                   34082	     32745 ns/op	   16529 B/op	     320 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/EncodedEventAsInput-4 	                   63714	     16536 ns/op	    7200 B/op	     190 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/RawEventAsInput-4     	                   70299	     17442 ns/op	    9037 B/op	     191 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/ComplexByteStreamAsInput-4         	   39494	     30623 ns/op	   13760 B/op	     310 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/ComplexEncodedEventAsInput-4       	   72664	     16951 ns/op	    7200 B/op	     180 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/ComplexRawEventAsInput-4           	   74628	     16978 ns/op	    7248 B/op	     181 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Output/SimpleEvent-4                     	  123708	      9405 ns/op	    4800 B/op	     110 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Output/ComplexEvent-4                    	   21565	     52149 ns/op	   35168 B/op	     510 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Output/ComplexLoggerSimpleEvent-4        	  116850	      9524 ns/op	    4800 B/op	     100 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Output/ComplexLoggerComplexEvent-4       	   18538	     56825 ns/op	   35011 B/op	     500 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Print/Simple-4                           	   44434	     27804 ns/op	   12816 B/op	     311 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Print/Complex-4                          	   43009	     27962 ns/op	   12816 B/op	     301 allocs/op
BenchmarkLogger/Runtime/SimpleLoggerPrintCall-4                                 	  336416	      3191 ns/op	    1544 B/op	      37 allocs/op
BenchmarkLogger/Runtime/SimpleLoggerLogCall-4                                   	  395946	      3094 ns/op	    1504 B/op	      35 allocs/op
BenchmarkLogger/Runtime/SimpleLoggerWriteString-4                               	  366129	      3421 ns/op	    1648 B/op	      38 allocs/op
BenchmarkLogger/Runtime/SimpleLoggerWriteEvent-4                                	  309572	      3977 ns/op	    1796 B/op	      43 allocs/op
BenchmarkLogger/Runtime/ComplexLoggerPrintCall-4                                	  283418	      3653 ns/op	    1686 B/op	      41 allocs/op
BenchmarkLogger/Runtime/ComplexLoggerLogCall-4                                  	  371646	      3299 ns/op	    1646 B/op	      39 allocs/op
BenchmarkLogger/Runtime/ComplexLoggerWriteString-4                              	  308035	      3603 ns/op	    1790 B/op	      42 allocs/op
BenchmarkLogger/Runtime/ComplexLoggerWriteEvent-4                               	  247177	      4410 ns/op	    1938 B/op	      47 allocs/op
PASS
coverage: [no statements]
ok  	github.com/zalgonoise/zlog/benchmark	59.513s
```

#### [`vendor_test.go`](./vendor_test.go)

```
Running tool: /usr/bin/go test -benchmem -run=^$ -coverprofile=/tmp/vscode-goqXOu1q/go-code-cover -bench ^BenchmarkVendorLoggers$ github.com/zalgonoise/zlog/benchmark

goos: linux
goarch: amd64
pkg: github.com/zalgonoise/zlog/benchmark
cpu: Intel(R) Core(TM) i5-4300M CPU @ 2.60GHz
BenchmarkVendorLoggers/Writing/SimpleText/ZeroLogger-4         	 7841997	     148.6 ns/op	     119 B/op	       0 allocs/op
BenchmarkVendorLoggers/Writing/SimpleText/StdLibLogger-4       	 5129163	     237.8 ns/op	      24 B/op	       1 allocs/op
BenchmarkVendorLoggers/Writing/SimpleText/ZapLogger-4          	 1448107	     835.3 ns/op	      64 B/op	       3 allocs/op
BenchmarkVendorLoggers/Writing/SimpleText/ZlogLogger-4         	 1681191	     704.4 ns/op	     368 B/op	       9 allocs/op
BenchmarkVendorLoggers/Writing/SimpleText/LogrusLogger-4       	  535584	      2287 ns/op	     480 B/op	      15 allocs/op
BenchmarkVendorLoggers/Writing/SimpleJSON/ZeroLogger-4         	10841835	     153.9 ns/op	      99 B/op	       0 allocs/op
BenchmarkVendorLoggers/Writing/SimpleJSON/ZapLogger-4          	 1653790	     697.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkVendorLoggers/Writing/SimpleJSON/ZlogLogger-4         	  824142	      1536 ns/op	     376 B/op	       6 allocs/op
BenchmarkVendorLoggers/Writing/SimpleJSON/LogrusLogger-4       	  435813	      2739 ns/op	    1080 B/op	      22 allocs/op
BenchmarkVendorLoggers/Writing/ComplexText/ZeroLogger-4        	  738760	      1552 ns/op	     288 B/op	      11 allocs/op
BenchmarkVendorLoggers/Writing/ComplexText/ZapLogger-4         	  328557	      3845 ns/op	     848 B/op	      21 allocs/op
BenchmarkVendorLoggers/Writing/ComplexText/ZlogLogger-4        	  227587	      5494 ns/op	    3756 B/op	      50 allocs/op
BenchmarkVendorLoggers/Writing/ComplexText/LogrusLogger-4      	  139341	      7770 ns/op	    2168 B/op	      43 allocs/op
BenchmarkVendorLoggers/Writing/ComplexJSON/ZeroLogger-4        	  760225	      1535 ns/op	     288 B/op	      11 allocs/op
BenchmarkVendorLoggers/Writing/ComplexJSON/ZapLogger-4         	  340016	      3539 ns/op	     784 B/op	      18 allocs/op
BenchmarkVendorLoggers/Writing/ComplexJSON/ZlogLogger-4        	  186678	      5911 ns/op	    2680 B/op	      40 allocs/op
BenchmarkVendorLoggers/Writing/ComplexJSON/LogrusLogger-4      	  202590	      5962 ns/op	    2592 B/op	      44 allocs/op
BenchmarkVendorLoggers/Init/SimpleText/ZeroLogger-4            	62708134	     18.23 ns/op	       0 B/op	       0 allocs/op
BenchmarkVendorLoggers/Init/SimpleText/StdLibLogger-4        1000000000	        0.7488 ns/op	       0 B/op	       0 allocs/op
BenchmarkVendorLoggers/Init/SimpleText/ZapLogger-4             	 1674553	     696.9 ns/op	    1560 B/op	      10 allocs/op
BenchmarkVendorLoggers/Init/SimpleText/ZlogLogger-4            	 2521554	     460.0 ns/op	     360 B/op	       9 allocs/op
BenchmarkVendorLoggers/Init/SimpleText/LogrusLogger-4          	 4365510	     250.1 ns/op	     368 B/op	       4 allocs/op
BenchmarkVendorLoggers/Init/SimpleJSON/ZeroLogger-4            	55591246	     18.11 ns/op	       0 B/op	       0 allocs/op
BenchmarkVendorLoggers/Init/SimpleJSON/ZapLogger-4             	 1436919	     713.3 ns/op	    1560 B/op	      10 allocs/op
BenchmarkVendorLoggers/Init/SimpleJSON/ZlogLogger-4            	 3499022	     379.9 ns/op	     296 B/op	       6 allocs/op
BenchmarkVendorLoggers/Init/SimpleJSON/LogrusLogger-4          	 4081352	     291.9 ns/op	     336 B/op	       4 allocs/op
BenchmarkVendorLoggers/Init/ComplexText/ZeroLogger-4           	64303657	     17.94 ns/op	       0 B/op	       0 allocs/op
BenchmarkVendorLoggers/Init/ComplexText/ZapLogger-4            	 1582852	     807.7 ns/op	    1560 B/op	      10 allocs/op
BenchmarkVendorLoggers/Init/ComplexText/ZlogLogger-4           	 2462248	     488.0 ns/op	     392 B/op	      10 allocs/op
BenchmarkVendorLoggers/Init/ComplexText/LogrusLogger-4         	 4343475	     265.0 ns/op	     368 B/op	       4 allocs/op
BenchmarkVendorLoggers/Init/ComplexJSON/ZeroLogger-4           	70297430	     16.92 ns/op	       0 B/op	       0 allocs/op
BenchmarkVendorLoggers/Init/ComplexJSON/ZapLogger-4            	 1000000	      1062 ns/op	    1560 B/op	      10 allocs/op
BenchmarkVendorLoggers/Init/ComplexJSON/ZlogLogger-4           	 3058534	     410.8 ns/op	     328 B/op	       7 allocs/op
BenchmarkVendorLoggers/Init/ComplexJSON/LogrusLogger-4         	 4698788	     265.4 ns/op	     336 B/op	       4 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleText/ZeroLogger-4         	 6653337	     183.9 ns/op	      16 B/op	       1 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleText/StdLibLogger-4       	 2795467	     431.3 ns/op	     224 B/op	       6 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleText/ZapLogger-4          	  721263	      1805 ns/op	    1624 B/op	      13 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleText/ZlogLogger-4         	  992654	      1277 ns/op	     728 B/op	      18 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleText/LogrusLogger-4       	  372478	      3219 ns/op	    1605 B/op	      29 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleJSON/ZeroLogger-4         	 6446944	     182.3 ns/op	      16 B/op	       1 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleJSON/ZapLogger-4          	  806187	      1686 ns/op	    1560 B/op	      10 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleJSON/ZlogLogger-4         	  616075	      1887 ns/op	     672 B/op	      12 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleJSON/LogrusLogger-4       	  340449	      3435 ns/op	    2127 B/op	      29 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexText/ZeroLogger-4        	  755805	      1562 ns/op	     304 B/op	      12 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexText/ZapLogger-4         	  264764	      4582 ns/op	    2408 B/op	      31 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexText/ZlogLogger-4        	  203493	      6133 ns/op	    4149 B/op	      60 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexText/LogrusLogger-4      	  118820	      8934 ns/op	    3291 B/op	      57 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexJSON/ZeroLogger-4        	  749808	      1582 ns/op	     304 B/op	      12 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexJSON/ZapLogger-4         	  260050	      4504 ns/op	    2344 B/op	      28 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexJSON/ZlogLogger-4        	  165524	      6508 ns/op	    3008 B/op	      47 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexJSON/LogrusLogger-4      	  149415	      6964 ns/op	    3644 B/op	      51 allocs/op
PASS
coverage: [no statements]
ok  	github.com/zalgonoise/zlog/benchmark	76.367s
```

_______________________

### 2022-08-10 - AMD Ryzen 3 PRO 3300U

#### [`logger_test.go`](./logger_test.go)

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