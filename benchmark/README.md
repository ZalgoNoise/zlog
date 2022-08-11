### 2022-08-11 - Intel i5-4300M CPU @ 2.60GHz

#### [`logger_test.go`](./logger_test.go)

```
Running tool: /usr/bin/go test -benchmem -run=^$ -coverprofile=/tmp/vscode-goBfzqxM/go-code-cover -bench ^BenchmarkLogger$ github.com/zalgonoise/zlog/benchmark

goos: linux
goarch: amd64
pkg: github.com/zalgonoise/zlog/benchmark
cpu: Intel(R) Core(TM) i5-4300M CPU @ 2.60GHz
BenchmarkLogger/Events/NewSimpleEvent-4         	                                  449596	      2326 ns/op	     760 B/op	      18 allocs/op
BenchmarkLogger/Events/NewSimpleEventWithLevel-4         	                          647486	      2186 ns/op	     760 B/op	      18 allocs/op
BenchmarkLogger/Events/NewComplexEvent-4                 	                           58184	     20662 ns/op	    3504 B/op	      98 allocs/op
BenchmarkLogger/Events/NewComplexEventWithCallStack-4    	                            5884	    243497 ns/op	   42675 B/op	     817 allocs/op
BenchmarkLogger/Formats/TextSimplest-4                   	                          436119	      3332 ns/op	    1186 B/op	      27 allocs/op
BenchmarkLogger/Formats/TextMostComplex-4                	                          372936	      3428 ns/op	    1421 B/op	      31 allocs/op
BenchmarkLogger/Formats/JSONCompact-4                    	                          349088	      3360 ns/op	    1135 B/op	      24 allocs/op
BenchmarkLogger/Formats/JSONIndented-4                   	                          206090	      5012 ns/op	    1631 B/op	      26 allocs/op
BenchmarkLogger/Formats/BSON-4                           	                          337112	      3481 ns/op	    1080 B/op	      24 allocs/op
BenchmarkLogger/Formats/CSV-4                            	                          277116	      4404 ns/op	    5120 B/op	      25 allocs/op
BenchmarkLogger/Formats/XML-4                            	                          136898	      9606 ns/op	    5704 B/op	      33 allocs/op
BenchmarkLogger/Formats/Gob-4                            	                          128773	     10904 ns/op	    3440 B/op	      71 allocs/op
BenchmarkLogger/Formats/Protobuf-4                       	                          481822	      2526 ns/op	     876 B/op	      21 allocs/op
BenchmarkLogger/Logger/Init/NewDefaultLogger-4           	                         4293262	     351.8 ns/op	     232 B/op	       4 allocs/op
BenchmarkLogger/Logger/Init/NewLoggerWithConfig-4        	                         1815289	     623.4 ns/op	     364 B/op	       9 allocs/op
BenchmarkLogger/Logger/Writing/Write/ByteStreamAsInput-4 	                          378957	      3158 ns/op	    1376 B/op	      32 allocs/op
BenchmarkLogger/Logger/Writing/Write/EncodedEventAsInput-4         	                  593271	      1757 ns/op	     720 B/op	      19 allocs/op
BenchmarkLogger/Logger/Writing/Write/RawEventAsInput-4             	                  564319	      2291 ns/op	     784 B/op	      20 allocs/op
BenchmarkLogger/Logger/Writing/Output/SimpleEvent-4                	                 1212639	      1015 ns/op	     770 B/op	      11 allocs/op
BenchmarkLogger/Logger/Writing/Output/ComplexEvent-4               	                  233558	      5250 ns/op	    3517 B/op	      51 allocs/op
BenchmarkLogger/Logger/Writing/Print/SimpleLogger-4                	                  401962	      2864 ns/op	    1296 B/op	      32 allocs/op
BenchmarkLogger/Logger/Writing/Print/ComplexLogger-4               	                  415329	      3160 ns/op	    1326 B/op	      32 allocs/op
BenchmarkLogger/MultiloggerX10/Init/NewDefaultLogger-4             	                  391958	      2847 ns/op	    2504 B/op	      42 allocs/op
BenchmarkLogger/MultiloggerX10/Init/NewLoggerWithConfig-4          	                  224793	      5949 ns/op	    3824 B/op	      92 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/ByteStreamAsInput-4   	                   32576	     35284 ns/op	   16172 B/op	     320 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/EncodedEventAsInput-4 	                   60885	     21385 ns/op	    8921 B/op	     190 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/RawEventAsInput-4     	                   61255	     20539 ns/op	    7264 B/op	     191 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/ComplexByteStreamAsInput-4         	   31292	     35774 ns/op	   13760 B/op	     310 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/ComplexEncodedEventAsInput-4       	   61741	     18796 ns/op	    7200 B/op	     180 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Write/ComplexRawEventAsInput-4           	   59116	     19393 ns/op	    7264 B/op	     181 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Output/SimpleEvent-4                     	  120268	     11291 ns/op	    4973 B/op	     110 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Output/ComplexEvent-4                    	   18633	     65547 ns/op	   35170 B/op	     510 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Output/ComplexLoggerSimpleEvent-4        	  131013	      9857 ns/op	    6240 B/op	     100 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Output/ComplexLoggerComplexEvent-4       	   21962	     53552 ns/op	   35009 B/op	     500 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Print/Simple-4                           	   39532	     35464 ns/op	   12816 B/op	     311 allocs/op
BenchmarkLogger/MultiloggerX10/Writing/Print/Complex-4                          	   37172	     30283 ns/op	   12816 B/op	     301 allocs/op
BenchmarkLogger/Runtime/SimpleLoggerPrintCall-4                                 	  363476	      3463 ns/op	    1544 B/op	      37 allocs/op
BenchmarkLogger/Runtime/SimpleLoggerLogCall-4                                   	  362574	      3625 ns/op	    1504 B/op	      35 allocs/op
BenchmarkLogger/Runtime/SimpleLoggerWriteString-4                               	  314961	      4043 ns/op	    1648 B/op	      38 allocs/op
BenchmarkLogger/Runtime/SimpleLoggerWriteEvent-4                                	  252272	      4103 ns/op	    1795 B/op	      43 allocs/op
BenchmarkLogger/Runtime/ComplexLoggerPrintCall-4                                	  306454	      3492 ns/op	    1686 B/op	      41 allocs/op
BenchmarkLogger/Runtime/ComplexLoggerLogCall-4                                  	  340654	      3601 ns/op	    1646 B/op	      39 allocs/op
BenchmarkLogger/Runtime/ComplexLoggerWriteString-4                              	  328732	      3936 ns/op	    1790 B/op	      42 allocs/op
BenchmarkLogger/Runtime/ComplexLoggerWriteEvent-4                               	  276338	      4430 ns/op	    1938 B/op	      47 allocs/op
PASS
coverage: [no statements]
ok  	github.com/zalgonoise/zlog/benchmark	67.772s
```

#### [`vendor_test.go`](./vendor_test.go)

```
Running tool: /usr/bin/go test -benchmem -run=^$ -coverprofile=/tmp/vscode-goBfzqxM/go-code-cover -bench ^BenchmarkVendorLoggers$ github.com/zalgonoise/zlog/benchmark

goos: linux
goarch: amd64
pkg: github.com/zalgonoise/zlog/benchmark
cpu: Intel(R) Core(TM) i5-4300M CPU @ 2.60GHz
BenchmarkVendorLoggers/Writing/SimpleText/ZeroLogger-4         	 7527403	     211.3 ns/op	      65 B/op	       0 allocs/op
BenchmarkVendorLoggers/Writing/SimpleText/StdLibLogger-4       	 3903853	     257.6 ns/op	      24 B/op	       1 allocs/op
BenchmarkVendorLoggers/Writing/SimpleText/ZapLogger-4          	 1352758	     848.5 ns/op	      64 B/op	       3 allocs/op
BenchmarkVendorLoggers/Writing/SimpleText/ZlogLogger-4         	 1644787	     780.9 ns/op	     368 B/op	       9 allocs/op
BenchmarkVendorLoggers/Writing/SimpleText/LogrusLogger-4       	  508026	      2288 ns/op	     480 B/op	      15 allocs/op
BenchmarkVendorLoggers/Writing/SimpleJSON/ZeroLogger-4         	10107571	     137.9 ns/op	      97 B/op	       0 allocs/op
BenchmarkVendorLoggers/Writing/SimpleJSON/ZapLogger-4          	 1734963	     678.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkVendorLoggers/Writing/SimpleJSON/ZlogLogger-4         	  831488	      1529 ns/op	     376 B/op	       6 allocs/op
BenchmarkVendorLoggers/Writing/SimpleJSON/LogrusLogger-4       	  396267	      2949 ns/op	    1208 B/op	      23 allocs/op
BenchmarkVendorLoggers/Writing/ComplexText/ZeroLogger-4        	  637591	      1717 ns/op	     416 B/op	      12 allocs/op
BenchmarkVendorLoggers/Writing/ComplexText/ZapLogger-4         	  298084	      4061 ns/op	    1104 B/op	      23 allocs/op
BenchmarkVendorLoggers/Writing/ComplexText/ZlogLogger-4        	  218557	      5960 ns/op	    3757 B/op	      50 allocs/op
BenchmarkVendorLoggers/Writing/ComplexText/LogrusLogger-4      	  116722	      9090 ns/op	    2424 B/op	      45 allocs/op
BenchmarkVendorLoggers/Writing/ComplexJSON/ZeroLogger-4        	  652778	      1669 ns/op	     416 B/op	      12 allocs/op
BenchmarkVendorLoggers/Writing/ComplexJSON/ZapLogger-4         	  302596	      3830 ns/op	    1040 B/op	      20 allocs/op
BenchmarkVendorLoggers/Writing/ComplexJSON/ZlogLogger-4        	  177649	      5922 ns/op	    2936 B/op	      42 allocs/op
BenchmarkVendorLoggers/Writing/ComplexJSON/LogrusLogger-4      	  159519	      6734 ns/op	    2976 B/op	      47 allocs/op
BenchmarkVendorLoggers/Init/SimpleText/ZeroLogger-4            137023033	     8.639 ns/op	       0 B/op	       0 allocs/op
BenchmarkVendorLoggers/Init/SimpleText/StdLibLogger-4         1000000000	    0.5567 ns/op	       0 B/op	       0 allocs/op
BenchmarkVendorLoggers/Init/SimpleText/ZapLogger-4             	 1564849	     779.2 ns/op	    1560 B/op	      10 allocs/op
BenchmarkVendorLoggers/Init/SimpleText/ZlogLogger-4            	 2549353	     485.7 ns/op	     336 B/op	       8 allocs/op
BenchmarkVendorLoggers/Init/SimpleText/LogrusLogger-4          	 4093740	     280.4 ns/op	     368 B/op	       4 allocs/op
BenchmarkVendorLoggers/Init/SimpleJSON/ZeroLogger-4            133774389	     8.972 ns/op	       0 B/op	       0 allocs/op
BenchmarkVendorLoggers/Init/SimpleJSON/ZapLogger-4             	 1365577	     801.9 ns/op	    1560 B/op	      10 allocs/op
BenchmarkVendorLoggers/Init/SimpleJSON/ZlogLogger-4            	 3185824	     377.7 ns/op	     296 B/op	       6 allocs/op
BenchmarkVendorLoggers/Init/SimpleJSON/LogrusLogger-4          	 4259244	     275.6 ns/op	     336 B/op	       4 allocs/op
BenchmarkVendorLoggers/Init/ComplexText/ZeroLogger-4           139552471	     8.675 ns/op	       0 B/op	       0 allocs/op
BenchmarkVendorLoggers/Init/ComplexText/ZapLogger-4            	 1475908	     807.6 ns/op	    1560 B/op	      10 allocs/op
BenchmarkVendorLoggers/Init/ComplexText/ZlogLogger-4           	 2585268	     467.8 ns/op	     368 B/op	       9 allocs/op
BenchmarkVendorLoggers/Init/ComplexText/LogrusLogger-4         	 4354825	     269.6 ns/op	     368 B/op	       4 allocs/op
BenchmarkVendorLoggers/Init/ComplexJSON/ZeroLogger-4           130574320	     9.335 ns/op	       0 B/op	       0 allocs/op
BenchmarkVendorLoggers/Init/ComplexJSON/ZapLogger-4            	 1427050	     818.8 ns/op	    1560 B/op	      10 allocs/op
BenchmarkVendorLoggers/Init/ComplexJSON/ZlogLogger-4           	 2606068	     415.6 ns/op	     328 B/op	       7 allocs/op
BenchmarkVendorLoggers/Init/ComplexJSON/LogrusLogger-4         	 3834078	     287.9 ns/op	     336 B/op	       4 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleText/ZeroLogger-4         	 6521924	     186.7 ns/op	      16 B/op	       1 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleText/StdLibLogger-4       	 2770024	     577.4 ns/op	     224 B/op	       6 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleText/ZapLogger-4          	  670376	      1737 ns/op	    1624 B/op	      13 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleText/ZlogLogger-4         	  957662	      1217 ns/op	     704 B/op	      17 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleText/LogrusLogger-4       	  331206	      3443 ns/op	    1600 B/op	      29 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleJSON/ZeroLogger-4         	 6781851	     174.8 ns/op	      16 B/op	       1 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleJSON/ZapLogger-4          	  754821	      1743 ns/op	    1560 B/op	      10 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleJSON/ZlogLogger-4         	  606476	      1983 ns/op	     672 B/op	      12 allocs/op
BenchmarkVendorLoggers/Runtime/SimpleJSON/LogrusLogger-4       	  281306	      3771 ns/op	    2263 B/op	      30 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexText/ZeroLogger-4        	  693561	      1833 ns/op	     432 B/op	      13 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexText/ZapLogger-4         	  251157	      5503 ns/op	    2664 B/op	      33 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexText/ZlogLogger-4        	  184062	      6336 ns/op	    4125 B/op	      59 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexText/LogrusLogger-4      	  108468	     11226 ns/op	    3541 B/op	      59 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexJSON/ZeroLogger-4        	  684825	      1757 ns/op	     432 B/op	      13 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexJSON/ZapLogger-4         	  228512	      4825 ns/op	    2600 B/op	      30 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexJSON/ZlogLogger-4        	  182956	      6593 ns/op	    3264 B/op	      49 allocs/op
BenchmarkVendorLoggers/Runtime/ComplexJSON/LogrusLogger-4      	  148375	      7316 ns/op	    4028 B/op	      54 allocs/op
PASS
coverage: [no statements]
ok  	github.com/zalgonoise/zlog/benchmark	83.133s
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