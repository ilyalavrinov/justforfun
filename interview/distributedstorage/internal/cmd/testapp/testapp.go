package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const (
	kb = 1024
	mb = 1024 * kb
	gb = 1024 * mb
)

func main() {
	argApiHost := flag.String("apihost", "localhost:7001", "apiservice to connect to")
	for _, filesize := range []int{1, 5, 6, 7, kb, mb, gb, 30 * gb} {
		err := runTest(filesize, *argApiHost)
		if err != nil {
			slog.Error("test failed", "filesize", filesize, "err", err)
			os.Exit(3)
		}
	}
}

func runTest(filesize int, connectTo string) error {
	slog.Info("new random data", "filesize", filesize)
	random, err := os.Open("/dev/random")
	if err != nil {
		return err
	}

	fileref := fmt.Sprintf("%d-%d", filesize, time.Now().UnixMicro())
	limitedReader := io.LimitReader(random, int64(filesize))
	md5er := md5.New()
	reader := io.TeeReader(limitedReader, md5er)
	url := fmt.Sprintf("http://%s/%s", connectTo, fileref)
	reqWrite, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		return err
	}
	reqWrite.ContentLength = int64(filesize)

	_, err = http.DefaultClient.Do(reqWrite)
	if err != nil {
		return err
	}

	md5sumOnSend := fmt.Sprintf("%x", md5er.Sum(nil))
	slog.Info("md5 for sending", "filesize", filesize, "md5", md5sumOnSend)

	md5er.Reset()
	reqRead, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	respRead, err := http.DefaultClient.Do(reqRead)
	if err != nil {
		return err
	}
	defer respRead.Body.Close()

	for {
		buffer := make([]byte, 1024*1024)
		readN, err := respRead.Body.Read(buffer)
		md5er.Write(buffer[:readN])
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
	}
	md5sumOnReceive := fmt.Sprintf("%x", md5er.Sum(nil))
	slog.Info("md5 for receiving", "filesize", filesize, "md5", md5sumOnReceive)

	if md5sumOnSend == md5sumOnReceive {
		slog.Info("md5 MATCHED", "md5_send", md5sumOnSend, "md5_receive", md5sumOnReceive)
	} else {
		return fmt.Errorf("md5 mismatch: send %q receive %q", md5sumOnSend, md5sumOnReceive)
	}

	return nil
}
