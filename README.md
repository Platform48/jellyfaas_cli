# JellyFaaS command line
Cli for JellyFaaS, user and admin tasks


To build:

```go build -o jellyfaas cmd/jellyfaas.go```
then run
```./jellyfaas```

or copy into a directory in your path

```cp jellyfaas /usr/local/bin```

then run

```jellyfaas```

for help

```./jellyfaas --help```

for example:

```
./jellyfaas user list
./jellyfaas user create -e|--email <email> -p|--password <password>
./jellyfaas secret --email <email> --password <password>
./jellyfaas deploy -z|--zip test.zip <-w|wait true>
./jellyfaas library
./jellyfaas library -d|--details <functionId>
```



