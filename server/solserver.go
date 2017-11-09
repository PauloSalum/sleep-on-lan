package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/julienschmidt/httprouter"
)

type Result struct {
	Message string `json:"message"`
}

const (
	baseLinux   = "systemctl"
	baseWindows = "shutdown -f"
)

var (
	shutdownFunc func(command string) error
	options      = []string{"suspend", "poweroff", "hibernate", "sleep"}
)

func init() {
	switch runtime.GOOS {
	case "windows":
		shutdownFunc = shutdownWindows
	case "linux":
		fmt.Println("###############\nPlease be sure that this script has sudo priviledges in order to run commands from this script\n################")
		shutdownFunc = shutdownLinux
	default:
		panic("Your os is not yet supported")
	}
}

func Start() {
	router := httprouter.New()
	router.POST("/:command", handleCommand)

	err := http.ListenAndServe(":7740", router)
	if err != nil {
		panic(err)
	}
}

func handleCommand(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	command := ps.ByName("command")
	enc := json.NewEncoder(w)
	if !stringInSlice(command, options) {
		handleError(w, errors.New("Option not available"), enc, http.StatusMethodNotAllowed)
		return
	}
	err := shutdownFunc(command)
	if err != nil {
		handleError(w, err, enc, http.StatusInternalServerError)
		return
	}
}

func handleError(w http.ResponseWriter, err error, enc *json.Encoder, code int) {
	w.WriteHeader(code)
	res := Result{Message: err.Error()}

	enc.Encode(&res)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func shutdownLinux(command string) error {
	var commandEx *exec.Cmd
	switch command {
	case "suspend":
		commandEx = exec.Command(baseLinux, "suspend")
	case "poweroff":
		commandEx = exec.Command(baseLinux, "poweroff")
	case "hibernate":
		commandEx = exec.Command(baseLinux, "hibernate")
	case "reboot": //Really?
		commandEx = exec.Command(baseLinux, "reboot")
	}
	return commandEx.Run()
}

func shutdownWindows(command string) error {
	var commandEx *exec.Cmd
	switch command {
	case "suspend":
		commandEx = exec.Command("rundll32", "powrprof.dll,SetSuspendState 0,1,0")
	case "poweroff":
		commandEx = exec.Command(baseWindows, "-s")
	case "hibernate":
		commandEx = exec.Command(baseWindows, "-h")
	case "reboot": //Really?
		commandEx = exec.Command(baseWindows, "-r")
	}
	return commandEx.Run()
}
