package collector

//go:generate /root/go/bin/bpf2go -cc clang connectProbe ./bpf/connect.c -- -I./bpf
//go:generate /root/go/bin/bpf2go -cc clang execveProbe ./bpf/execve.c -- -I./bpf
