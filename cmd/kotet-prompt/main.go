package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const fallback string = "$ "

var socket_file string = "/tmp/kotet-prompt.sock"

func main() {
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := net.Dial("unix", socket_file)
		if err != nil {
			log.Println(err.Error())
			return
		}
		defer conn.Close()
		err = conn.(*net.UnixConn).CloseWrite()
		if err != nil {
			log.Println(err.Error())
			return
		}
	}()

	return_code := flag.Int("return", 0, "return code")
	flag.Parse()
	var buf bytes.Buffer
	var clock, pwd, git, ret string
	wg.Add(1)
	go func() {
		clock = Clock()
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		pwd = Pwd()
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		git = Git()
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		ret = ReturnCode(*return_code)
		wg.Done()
	}()
	wg.Wait()
	if clock != "" {
		buf.WriteString(Color(clock, "97"))
	}
	if pwd != "" {
		buf.WriteByte(' ')
		buf.WriteString(Color(pwd, "96"))
	}
	if git != "" {
		buf.WriteByte(' ')
		buf.WriteString(Color(git, "95"))
	}
	if ret != "" {
		buf.WriteByte(' ')
		buf.WriteString(Color(ret, "91"))
	}
	fmt.Printf("\033[4m\033[1m%s\033[0m\n$ ", buf.String())
}

func Color(str string, code string) string {
	return "\033[" + code + "m[" + str + "]\033[4m\033[1m"
}

func ReturnCode(code int) string {
	if code == 0 {
		return ""
	}
	return fmt.Sprint(code)
}

func Clock() string {
	t := time.Now()
	return t.Format("01/02(Mon)15:04:05")
}

func Git() string {
	cmd, err := exec.Command("git", "symbolic-ref", "--short", "HEAD").Output()
	if err == nil {
		return strings.TrimSpace(string(cmd))
	}
	cmd, err = exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(cmd))
}

func Pwd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Println(err.Error())
		fmt.Println(fallback)
		os.Exit(1)
	}
	curr, err := user.Current()
	if err != nil {
		log.Println(err.Error())
		fmt.Println(fallback)
		os.Exit(1)
	}
	home := curr.HomeDir
	if home == wd {
		return "~"
	}
	if strings.HasPrefix(wd, home) {
		relpath, err := filepath.Rel(home, wd)
		if err != nil {
			log.Println(err.Error())
			fmt.Println(fallback)
			os.Exit(1)
		}
		return TrimPath("~/" + relpath)
	}
	return TrimPath(wd)
}

func TrimPath(path string) string {
	maxlength := 50
	if len(path) < maxlength {
		return path
	}
	slist := strings.Split(path, "/")
	front := slist[0]
	back := slist[len(slist)-1]
	for i := 2; 0 < len(slist)-i && len(front)+len(back)+len(slist[len(slist)-i])+5 < maxlength; i += 1 {
		back = slist[len(slist)-i] + "/" + back
	}
	return front + "/.../" + back
}
