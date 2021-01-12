package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/gopaddle-io/sail/misc"
	trace "github.com/gopaddle-io/sail/trace"
	util "github.com/gopaddle-io/sail/util"
	"github.com/gopaddle-io/sail/util/cmd"

	flag "github.com/spf13/pflag"
)

func main() {

	listCommand := flag.NewFlagSet("list", flag.ExitOnError)
	dockerizeCommand := flag.NewFlagSet("dockerize", flag.ExitOnError)
	helpCommand := flag.NewFlagSet("help", flag.ExitOnError)
	listTextPtr := listCommand.StringP("all", "a", "", "list --all process")
	listHelpPtr := listCommand.BoolP("help", "h", false, "Help regarding commands")
	dockerizeTextPtr := dockerizeCommand.StringP("pid", "p", "", "pid of the process to trace")
	dockerizeTimePtr := dockerizeCommand.IntP("time", "t", 2, "Time in seconds to trace the process to build its docker profile")
	dockerizeImagePtr := dockerizeCommand.StringP("imageName", "i", "final", "Name of the final docker image")
	dockerizeHelpPtr := dockerizeCommand.BoolP("help", "h", false, "Help regarding commands")

	if len(os.Args) < 2 {
		fmt.Println("Command required.")
		fmt.Println("To check the commands possible, run sail --help or -h")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "list":
		listCommand.Parse(os.Args[2:])
	case "dockerize":
		dockerizeCommand.Parse(os.Args[2:])
	case "--help":
		helpCommand.Parse(os.Args[2:])
	case "-h":
		helpCommand.Parse(os.Args[2:])
	default:
		fmt.Println("Invalid command. Please check sail help for available options")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if listCommand.Parsed() {

		if *listHelpPtr == true {
			fmt.Print("List the Process running on the Machine for the current user. \n")
			fmt.Print("\n")
			fmt.Printf("   sail list (--all | -a) process.\n")
			os.Exit(0)
		}

		if *listTextPtr == "" {
			listCommand.PrintDefaults()
			os.Exit(1)
		}

		if len(os.Args) >= 3 && os.Args[3] == "process" {
			out, err := exec.Command("ps", "-eo", "pid,ppid,cmd").Output()
			if err != nil {
				fmt.Printf("%s", err)
			}
			output := string(out[:])
			fmt.Println(output)
		} else {
			fmt.Println("Enter complete command (try : sail list --all process)")
		}
	}

	if dockerizeCommand.Parsed() {

		if *dockerizeHelpPtr == true {
			fmt.Print("Migrate a running linux process in to a Docker Image. \n")
			fmt.Print("\n")
			fmt.Printf("sail dockerize --pid <process id> [--time <time in seconds>] [--imageName <docker image name>]\n")
			fmt.Print("\n")
			fmt.Printf("    %-20s%s\n", "-p, --pid", "pid of the process to trace.")
			fmt.Printf("    %-20s%s\n", "-t, --time", "Time in seconds to trace the process to build its docker profile. Defaults to 2 seconds.")
			fmt.Printf("    %-20s%s\n", "-i, --imageName", "Name of the final docker image. Defaults to 'final'.")
			os.Exit(0)
		}

		if *dockerizeTextPtr == "" {
			dockerizeCommand.PrintDefaults()
			os.Exit(1)
		}
		//values4 := map[string]string{"finalimagename": *dockerizeImagePtr,"home": "/tmp"}
		dir := []string{"/tmp/sail", "packages.fmt"}

		requestID := util.NewRequestID()
		defer func() {
			if r := recover(); r != nil {
				misc.PanicHandler(r, requestID)
			}
		}()

		// Start Tracing
		log.Println("start tracing...")
		if _, err := trace.StartTracing_noreq(*dockerizeTextPtr, *dockerizeTimePtr, requestID); err != nil {
			log.Println("tracing failed :", err.Error())
		}
		log.Println("tracing completed")

		// Docker Create
		_, osname, osver, _ := cmd.GetOS()
		log.Println("Docker creating...")
		if _, err := trace.DockerCreate_noreq(osname, osver, "final", requestID); err != nil {
			log.Println("Docker container creation failed :", err.Error())
		}
		log.Println("Docker creating completed")

		//Docker Copy
		log.Println("Docker file copying ...")
		if _, err := trace.DockerCopy_noreq(dir, requestID); err != nil {
			log.Println("Docker file copy failed :", err.Error())
		}
		log.Println("Docker file copying completed")
		log.Println("Copying fmt file of trace to container...")

		//Docker final image create
		trace.FinalImageCreate_noreq("$HOME", *dockerizeImagePtr, requestID)
		log.Println(*dockerizeImagePtr + " created")
		log.Println("To check the image, use command : docker image inspect " + *dockerizeImagePtr)

	}

	if helpCommand.Parsed() {
		fmt.Print("Migrate a running linux process in to a Docker Image. \n")
		fmt.Println("\nEnter sail <command> --help or -h for more details on specific commands and their arguments.")
		fmt.Println("\nCommands:")
		fmt.Printf("    %-12s%s\n", "dockerize", "Migrate a linux process to Docker Image")
		fmt.Printf("    %-12s%s\n", "list", "List all processes owned by the current user")
	}
}
