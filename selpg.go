package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	flag "github.com/ogier/pflag"
)

/*============================ types =============================*/

type selpgArgs struct {
	start_page  int
	end_page    int
	page_len    int
	page_type   string
	print_dest  string
	in_filename string
}

/*============================ main ===============================*/
func main() {
	sa := new(selpgArgs)

	//定义参数
	flag.IntVarP(&sa.start_page, "s", "s", -1, "the start page")
	flag.IntVarP(&sa.end_page, "e", "e", -1, "the end page")
	flag.IntVarP(&sa.page_len, "l", "l", 72, "the paging form")
	flag.StringVarP(&sa.print_dest, "d", "d", "", "the printer")

	exist_f := flag.Bool("f", false, "")

	flag.Parse()

	if *exist_f {
		sa.page_type = "f"
		sa.page_len = -1
	} else {
		sa.page_type = "l"
	}

	if flag.NArg() == 2 {
		sa.in_filename = flag.Arg(1)
		//fmt.Printf("%s", sa.in_filename)
	} else {
		sa.in_filename = ""
	}

	handle_args(*sa, flag.NArg())

	process_input(*sa)

}

func usage() {
	fmt.Fprintf(os.Stderr, "\nUSAGE: ./selpg [-s start_page] [-e end_page] [ -f | -l lines_per_page ] [ -ddest ] [ in_filename ]\n")
}

func handle_args(sa selpgArgs, notFlagNum int) {
	if sa.start_page > sa.end_page || sa.start_page < 1 {
		//fmt.Printf("%d %d", sa.end_page, sa.start_page)
		usage()
		os.Exit(1)
	}
	if notFlagNum != 2 && notFlagNum != 1 {
		//fmt.Printf("%d\n", notFlagNum)
		//fmt.Printf("2")
		usage()
		os.Exit(1)
	}
	if sa.page_type == "f" {
		if sa.page_len != -1 {
			fmt.Printf("3")
			usage()
			os.Exit(1)
		}
	}
}

func process_input(sa selpgArgs) {
	fin := os.Stdin
	var err error

	//set the input source
	if sa.in_filename != "" {
		fin, err = os.Open(sa.in_filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "selpg: could not open input file \"%s\"\n", sa.in_filename)
			fmt.Println(err)
			usage()
			os.Exit(1)
		}
		defer fin.Close()

	}

	//set the output
	fout := os.Stdout
	var inpipe io.WriteCloser
	if sa.print_dest != "" {
		cmd := exec.Command("lp", "-d"+sa.print_dest)
		inpipe, err = cmd.StdinPipe()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer inpipe.Close()
		cmd.Stdout = fout
		cmd.Start()
	}

	reader := bufio.NewReader(fin)
	page_ctr := 1
	line_ctr := 0
	//-l
	if sa.page_type == "l" {
		line := bufio.NewScanner(fin)
		for line.Scan() {
			if page_ctr >= sa.start_page && page_ctr <= sa.end_page {
				if sa.print_dest != "" {
					inpipe.Write([]byte(line.Text() + "\n"))
				} else {
					fout.Write([]byte(line.Text() + "\n"))
				}
			}
			line_ctr++
			if line_ctr > sa.page_len {
				page_ctr++
				line_ctr = 1
			}
		}
	} else {
		page_ctr = 1
		for {
			page, errr := reader.ReadString('\f')
			if errr != nil || errr == io.EOF {
				if errr == io.EOF {
					if page_ctr >= sa.start_page && page_ctr <= sa.end_page {
						fmt.Fprintf(fout, "%s", page)
					}
				}
				break
			}
			page = strings.Replace(page, "\f", "", -1)
			page_ctr++
			if page_ctr >= sa.start_page && page_ctr <= sa.end_page {
				fmt.Fprintf(fout, "%s", page)
			}
		}
	}

	if page_ctr < sa.end_page {
		fmt.Fprintf(os.Stderr, "./selpg: end_page (%d) greater than total pages (%d), less output than expected\n", sa.end_page, page_ctr)
	}

}
