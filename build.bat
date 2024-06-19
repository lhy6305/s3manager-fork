@echo off

set GOOS=windows
set GOARCH=amd64
go build -v -ldflags "-s -w" -trimpath -o s3manager_x64.exe .

set GOOS=windows
set GOARCH=386
go build -v -ldflags "-s -w" -trimpath -o s3manager_x86.exe .

set GOOS=linux
set GOARCH=amd64
go build -v -ldflags "-s -w" -trimpath -o s3manager_x64.elf .

set GOOS=linux
set GOARCH=386
go build -v -ldflags "-s -w" -trimpath -o s3manager_x86.elf .

set GOOS=linux
set GOARCH=mipsle
go build -v -ldflags "-s -w" -trimpath -o s3manager_mipsle32.elf .

set GOOS=linux
set GOARCH=mips64le
go build -v -ldflags "-s -w" -trimpath -o s3manager_mipsle64.elf .

set GOOS=linux
set GOARCH=arm
go build -v -ldflags "-s -w" -trimpath -o s3manager_arm.elf .

set GOOS=linux
set GOARCH=arm64
go build -v -ldflags "-s -w" -trimpath -o s3manager_arm64.elf .
