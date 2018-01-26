# Benchmark

CPU & RAM:

```
$ lscpu
Architecture:        x86_64
CPU op-mode(s):      32-bit, 64-bit
Byte Order:          Little Endian
CPU(s):              4
On-line CPU(s) list: 0-3
Thread(s) per core:  2
Core(s) per socket:  2
Socket(s):           1
NUMA node(s):        1
Vendor ID:           GenuineIntel
CPU family:          6
Model:               58
Model name:          Intel(R) Core(TM) i5-3210M CPU @ 2.50GHz
Stepping:            9
CPU MHz:             1214.294
CPU max MHz:         3100.0000
CPU min MHz:         1200.0000
BogoMIPS:            4988.41
Virtualization:      VT-x
L1d cache:           32K
L1i cache:           32K
L2 cache:            256K
L3 cache:            3072K
NUMA node0 CPU(s):   0-3
Flags:               fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx rdtscp lm constant_tsc arch_perfmon pebs bts rep_good nopl xtopology nonstop_tsc aperfmperf eagerfpu pni pclmulqdq dtes64 monitor ds_cpl vmx est tm2 ssse3 cx16 xtpr pdcm pcid sse4_1 sse4_2 x2apic popcnt tsc_deadline_timer aes xsave avx f16c rdrand lahf_lm epb retpoline kaiser tpr_shadow vnmi flexpriority ept vpid fsgsbase smep erms xsaveopt dtherm ida arat pln pts

$ free -h
              total        used        free      shared  buff/cache   available
Mem:            11G        536M        9.0G         11M        2.2G         10G
Swap:          9.6G          0B        9.6G
```

## systemd service files & nginx conf

- nginx that service pages

```
$ cat /usr/lib/systemd/system/nginx.service
[Unit]
Description=A high performance web server and a reverse proxy server
After=network.target network-online.target nss-lookup.target

[Service]
Type=forking
PIDFile=/run/nginx.pid
PrivateDevices=yes
SyslogLevel=err
User=root

LimitNOFILE=4096000
ExecStart=/usr/bin/nginx -g 'pid /run/nginx.pid; error_log stderr;'
ExecReload=/usr/bin/nginx -s reload
KillSignal=SIGQUIT
KillMode=mixed

[Install]
WantedBy=multi-user.target

$ cat /etc/nginx/nginx.conf
worker_processes 2;
worker_rlimit_nofile 204800;

error_log  /var/log/nginx/error.log;

events {
    worker_connections  10240;
    use epoll;
    multi_accept on;
}

http {
    include       mime.types;
    default_type  application/text-plain;

    access_log  /var/log/nginx/access.log;

    sendfile        on;

    keepalive_timeout  65;

    gzip  on;

    server {
        listen       80;
        server_name  localhost;

        location / {
            root /usr/share/nginx/html;
            index index.html;
        }

        location /hello {
            return 200 'hello world';
        }
    }
}
```

- nginx as proxy

```
$ cat /etc/systemd/system/nginx_proxy.service
[Unit]
Description=A high performance web server and a reverse proxy server
After=network.target network-online.target nss-lookup.target

[Service]
Type=forking
PIDFile=/tmp/nginx.pid
PrivateDevices=yes
SyslogLevel=err

LimitNOFILE=4096000
ExecStart=/usr/bin/nginx -g 'pid /tmp/nginx.pid;' -c /tmp/nginx_proxy.conf
ExecReload=/usr/bin/nginx -s reload
KillSignal=SIGQUIT
KillMode=mixed

[Install]
WantedBy=multi-user.target

$ cat /tmp/nginx_proxy.conf 
worker_processes 2;
worker_rlimit_nofile 204800;

error_log  /var/log/nginx/error.log;

events {
    worker_connections  10240;
    use epoll;
    multi_accept on;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/text-plain;

    access_log  /var/log/nginx/access.log;

    sendfile        on;

    keepalive_timeout  65;

    gzip  on;

    server {
        listen       9999;
        server_name  localhost;

        location / {
            proxy_pass http://127.0.0.1:80/;
        }
    }
}
```

- guard

```
$ cat /etc/systemd/system/guard.service 
[Unit]
Description=Guard Circuit Breaker
After=network.target

[Service]
Type=simple
User=root
Environment=GOMAXPROCS=2
LimitNOFILE=204800
WorkingDirectory=/usr/local/bin/
ExecStart=/usr/local/bin/guard
Restart=on-abort

[Install]
WantedBy=multi-user.target

$ cat ~/backends.json 
{
    "name": "www.example.com",
    "backends": ["127.0.0.1:80", "127.0.0.1:80", "127.0.0.1:80"],
    "weights": [5, 1, 1],
    "ratio": 0.3,
    "paths": ["/", "/doc"],
    "methods": ["GET", "GET"]
}
```

## Start!

- make sure all things goes well

```bash
$ curl -X GET --header "Host: www.example.com" -I http://127.0.0.1:80/
HTTP/1.1 200 OK
Server: nginx/1.12.2
Date: Fri, 26 Jan 2018 02:02:56 GMT
Content-Type: text/html
Content-Length: 612
Last-Modified: Wed, 22 Nov 2017 19:48:42 GMT
Connection: keep-alive
ETag: "5a15d49a-264"
Accept-Ranges: bytes

$ curl -X GET --header "Host: www.example.com" -I http://127.0.0.1:9999/
HTTP/1.1 200 OK
Server: nginx/1.12.2
Date: Fri, 26 Jan 2018 02:03:03 GMT
Content-Type: text/html
Content-Length: 612
Connection: keep-alive
Last-Modified: Wed, 22 Nov 2017 19:48:42 GMT
ETag: "5a15d49a-264"
Accept-Ranges: bytes

$ curl -X GET --header "Host: www.example.com" -I http://127.0.0.1:23456/
HTTP/1.1 200 OK
Server: nginx/1.12.2
Date: Fri, 26 Jan 2018 02:03:07 GMT
Content-Type: text/html
Content-Length: 612
Last-Modified: Wed, 22 Nov 2017 19:48:42 GMT
Etag: "5a15d49a-264"
Accept-Ranges: bytes
```

- test nginx server itself

```
$ wrk --latency -H "Host: www.example.com" -c 2048 -d 30 -t 2 http://127.0.0.1:80
Running 30s test @ http://127.0.0.1:80
  2 threads and 2048 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    30.71ms   15.74ms 274.65ms   73.86%
    Req/Sec    16.34k     2.45k   22.59k    77.59%
  Latency Distribution
     50%   30.73ms
     75%   35.13ms
     90%   57.45ms
     99%   65.53ms
  973250 requests in 30.09s, 788.89MB read
  Socket errors: connect 1029, read 0, write 0, timeout 0
Requests/sec:  32343.50
Transfer/sec:     26.22MB
```

- test nginx proxy

```
$ wrk --latency -H "Host: www.example.com" -c 2048 -d 30 -t 2 http://127.0.0.1:9999
Running 30s test @ http://127.0.0.1:9999
  2 threads and 2048 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    90.05ms   94.04ms   1.23s    97.98%
    Req/Sec     6.04k     1.99k    9.98k    56.38%
  Latency Distribution
     50%   81.12ms
     75%  134.20ms
     90%  147.84ms
     99%  527.96ms
  358981 requests in 30.04s, 290.98MB read
  Socket errors: connect 1029, read 0, write 0, timeout 110
Requests/sec:  11951.02
Transfer/sec:      9.69MB
```

- test guard

```
$ wrk --latency -H "Host: www.example.com" -c 2048 -d 30 -t 2 http://127.0.0.1:23456
Running 30s test @ http://127.0.0.1:23456
  2 threads and 2048 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    43.38ms   21.79ms   1.10s    91.83%
    Req/Sec    11.83k     2.45k   16.43k    63.70%
  Latency Distribution
     50%   42.37ms
     75%   48.88ms
     90%   58.47ms
     99%   77.49ms
  704433 requests in 30.07s, 554.91MB read
  Socket errors: connect 1029, read 0, write 0, timeout 0
Requests/sec:  23425.98
Transfer/sec:     18.45MB
```

## FAQ

- why not set CPU affinity?

    for fair. guard does set this. but you can set CPU affinity by using `taskset` for guard. your benchmark is welcome!
