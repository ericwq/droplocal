# DropLocal

allow LAN user transfer files to each other without knowing machine name / port number first. it's based on zero configuration network (or bonjour) technology. the original idea is to share document/photo between linux/windows/mobile devices.

currently, it's a command line tool. it has been tested on the Mac laptop. I plan to add a gui for it.

## install
` go get github.com/ericwq/droplocal ` get the source and build the executable file by ` go build `
## usage

### step1: start the droplocal server
1. find the executive file and run ` ./droplocal -s & `
2. optionally you can also redirect the log to somewhere e.g. 
```
./droplocal -s & > /tmp/droplocal.2020.06.log
```
you can start multiple servers on you local LAN. here is the output from the above command
```
/Users/qiwang/dev/droplocal
qiwang@Einstein droplocal % ./droplocal -s 
2020/07/06 15:01:22 the serve dir is /Users/qiwang
2020/07/06 15:01:22 username admin, password password
2020/07/06 15:01:22 using the name:Drop Local 27422438
2020/07/06 15:01:22   Drop Local 27422438 listening on 2121
2020/07/06 15:01:22 advertise the Drop Local 27422438 on local link`
```
you can also stop the server via Ctrl-C. or just use the *kill* command
```
qiwang@Einstein droplocal % ./droplocal -s 
2020/07/06 15:01:22 the serve dir is /Users/qiwang
2020/07/06 15:01:22 username admin, password password
2020/07/06 15:01:22 using the name:Drop Local 27422438
2020/07/06 15:01:22   Drop Local 27422438 listening on 2121
2020/07/06 15:01:22 advertise the Drop Local 27422438 on local link
^C2020/07/06 15:03:43 Error starting server:ftp: Server closed
2020/07/06 15:03:43 shutting down...
qiwang@Einstein droplocal % 
```
### step2: start the droplocal client
1. run with `./droplocal -h`, you will know all the command parameters. 
- `-s` means run in the server mode, 
- `-d` specified the directory which you will get the droped files
- `-u` specified the username
- `-p` specified the user passward
- `-f` specified the file you want to transfer
```
qiwang@Einstein droplocal % ./droplocal -h
Usage of ./droplocal:
  -d string
    	dir to serve file uploaded
  -f string
    	source file for upload
  -p string
    	Password for login (default "password")
  -s	run in server mode
  -u string
    	Username for login (default "admin")
```
2. run with `./droplocal -f util.go` , you will get the following output
```
qiwang@Einstein droplocal % ./droplocal -f util.go
index |  service name @ machine
[  0] | Drop\ Local\ 3949183984@Oppenheimer.local.local.
please choose the destination, please use the index to choose.
0
mission accomplished!
qiwang@Einstein droplocal % 
```
- the system will query the LAN, find the available drop local services, here only one service is available.
- the system prompt you to choose the target service/machine, in this case, only "Drop\ Local\ 3949183984" at machine "Oppenheimer.local" is available
- use the index to choose the machine, and return to confirm the choice.
- the system will transfer the file to the target machine.

now. the util.go file has been transfered to the Oppenheimer machine on the LAN. Of course, A lives droplocal server is running on that machine. 
### step3: check the user's home directory on Oppenheimer
you will find a util.go file exist on that machine.
