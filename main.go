/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : main.go

* Purpose :

* Creation Date : 01-15-2014

* Last Modified : Fri 16 May 2014 10:52:14 PM UTC

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kiyor/gourl/lib"
	"github.com/kiyor/nagiosToJson"
	"github.com/wsxiaoys/terminal/color"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var (
	statusf       *string = flag.String("f", "/usr/local/nagios/var/status.dat", "status file")
	all           *bool   = flag.Bool("all", false, "get all info")
	mute          *bool   = flag.Bool("mute", false, "enable show mute info")
	mutehost      *string = flag.String("mutehost", "", "mute all by host")
	muteservice   *string = flag.String("muteservice", "", "mute all by service")
	unmutehost    *string = flag.String("unmutehost", "", "unmute all by host")
	unmuteservice *string = flag.String("unmuteservice", "", "unmute all by service")
	e             *bool   = flag.Bool("exec", false, "toggle action")
	cleanmute     *bool   = flag.Bool("cleanmute", false, "show all mute info")
	ack           *bool   = flag.Bool("ack", false, " enable show ack")
	url           *string = flag.String("url", "", "get status file by url")
	cmdfile       *string = flag.String("cmdfile", "/usr/local/nagios/var/rw/nagios.cmd", "custom cmd file")
)

func init() {
	flag.Parse()
	if *url != "" {
		setStatByUrl(*url)
	} else {
		nagiosToJson.SetStatFile(*statusf)
	}
}

// this is for sometime you testing status.dat file on local but need newest file
func setStatByUrl(url string) {
	var req gourl.Req
	req.Url = url
	res, err := req.GetString()
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = ioutil.WriteFile("/tmp/temp.dat", []byte(res), 0644)
	if err != nil {
		log.Fatalln(err.Error())
	}
	nagiosToJson.SetStatFile("/tmp/temp.dat")
}
func chkReErr(err error) {
	if err != nil {
		log.Fatalln(err.Error())
		os.Exit(1)
	}
}

// still a lot of duplicate code, need make it clean and might use single func just pass nagios api command
func main() {
	var stat nagiosToJson.Mainstat
	a := nagiosToJson.GetStat()
	json.Unmarshal(a, &stat)
	if *all {
		j, _ := json.MarshalIndent(stat, "", "    ")
		fmt.Println(string(j))
	} else {
		create := str2time(stat.Info.Created)
		fmt.Println("data Created from", time.Since(create), "ago")
		title()
		var reh, res *regexp.Regexp
		var err error
		if *mutehost != "" {
			reh, err = regexp.Compile(`.*` + *mutehost + `.*`)
			chkReErr(err)
		} else {
			reh, err = regexp.Compile(`.*` + *unmutehost + `.*`)
			chkReErr(err)
		}
		if *muteservice != "" {
			res, err = regexp.Compile(`.*` + *muteservice + `.*`)
			chkReErr(err)
		} else {
			res, err = regexp.Compile(`.*` + *unmuteservice + `.*`)
			chkReErr(err)
		}
		// for loop all element by hostname
		for hostname, v := range stat.Hoststatus {
			// main if statement for flag
			if *cleanmute {
				// sub if statement for logic
				if !notifications(v) && state(v) == 0 && !acknowledged(v) && active(v) {
					output(hostname, "HOST", v)
					if *e {
						c := fmt.Sprintf("echo \"[%d] ENABLE_HOST_NOTIFICATIONS;%s\n\" > %s", time.Now().Unix(), hostname, *cmdfile)
						run(c)
					}
				}
			} else if *mutehost != "" {
				if reh.MatchString(hostname) && *muteservice == "" {
					output(hostname, "HOST", v)
					if *e {
						c := fmt.Sprintf("echo \"[%d] DISABLE_HOST_NOTIFICATIONS;%s\n\" > %s", time.Now().Unix(), hostname, *cmdfile)
						run(c)
						c = fmt.Sprintf("echo \"[%d] DISABLE_HOST_SVC_NOTIFICATIONS;%s\n\" > %s", time.Now().Unix(), hostname, *cmdfile)
						run(c)
					}
				}
			} else if *unmutehost != "" {
				if reh.MatchString(hostname) && !notifications(v) && state(v) == 0 && !acknowledged(v) && active(v) {
					output(hostname, "HOST", v)
					if *e {
						c := fmt.Sprintf("echo \"[%d] ENABLE_HOST_NOTIFICATIONS;%s\n\" > %s", time.Now().Unix(), hostname, *cmdfile)
						run(c)
					}
				}
			}

			// for loop all element by servicename
			for servicename, v2 := range v.Servicestatus {
				if *cleanmute {
					if !notifications(v2) && state(v2) == 0 && !acknowledged(v2) && active(v2) {
						output(hostname, servicename, v2)
						if *e {
							c := fmt.Sprintf("echo \"[%d] ENABLE_SVC_NOTIFICATIONS;%s;%s\n\" > %s", time.Now().Unix(), hostname, servicename, *cmdfile)
							run(c)
						}
					}
				} else if *muteservice != "" {
					// if more than one logic then use bool variable help
					var send bool
					if *mutehost != "" {
						if reh.MatchString(hostname) && res.MatchString(servicename) {
							send = true
						}
					}
					if (*mutehost == "" && res.MatchString(servicename)) || send { // if only related on service or host&service name all true, then output
						output(hostname, servicename, v2)
						if *e {
							c := fmt.Sprintf("echo \"[%d] DISABLE_SVC_NOTIFICATIONS;%s;%s\n\" > %s", time.Now().Unix(), hostname, servicename, *cmdfile)
							run(c)
						}
					}
				} else if *unmuteservice != "" {
					var send bool
					if *unmutehost != "" {
						if reh.MatchString(hostname) && res.MatchString(servicename) {
							send = true
						}
					}
					if (send || (*unmutehost == "" && res.MatchString(servicename))) && !notifications(v2) && state(v2) == 0 && !acknowledged(v2) && active(v2) { // if service name match and muted, state is good and not acked, active checked
						output(hostname, servicename, v2)
						if *e {
							c := fmt.Sprintf("echo \"[%d] ENABLE_SVC_NOTIFICATIONS;%s;%s\n\" > %s", time.Now().Unix(), hostname, servicename, *cmdfile)
							run(c)
						}

					}
				} else if *unmutehost != "" {
					if reh.MatchString(hostname) && !notifications(v2) && state(v2) == 0 && !acknowledged(v2) && active(v2) { // if hostname match and muted, state is good and not acked, active checked
						output(hostname, servicename, v2)
						if *e {
							c := fmt.Sprintf("echo \"[%d] ENABLE_SVC_NOTIFICATIONS;%s;%s\n\" > %s", time.Now().Unix(), hostname, servicename, *cmdfile)
							run(c)
						}
					}
				}
			}
		}
	}
}

// run os cmd
func run(c string) {
	// 	fmt.Println(c)
	cmd := exec.Command("/bin/bash", "-c", c)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalln(err.Error())
	}
}

// four useful readable func
func state(v interface{}) int {
	var res int
	switch v := v.(type) {
	default:
		return 0
	case *nagiosToJson.Hoststatus:
		res, _ = strconv.Atoi(v.Current_state)
		return res
	case *nagiosToJson.Servicestatus:
		res, _ = strconv.Atoi(v.Current_state)
		return res
	}
}
func active(v interface{}) bool {
	var res int
	switch v := v.(type) {
	default:
		return false
	case *nagiosToJson.Hoststatus:
		res, _ = strconv.Atoi(v.Active_checks_enabled)
		if res == 1 {
			return true
		}
	case *nagiosToJson.Servicestatus:
		res, _ = strconv.Atoi(v.Active_checks_enabled)
		if res == 1 {
			return true
		}
	}
	return false
}
func notifications(v interface{}) bool {
	var res int
	switch v := v.(type) {
	default:
		return false
	case *nagiosToJson.Hoststatus:
		res, _ = strconv.Atoi(v.Notifications_enabled)
		if res == 1 {
			return true
		}
	case *nagiosToJson.Servicestatus:
		res, _ = strconv.Atoi(v.Notifications_enabled)
		if res == 1 {
			return true
		}
	}
	return false
}
func acknowledged(v interface{}) bool {
	var res int
	switch v := v.(type) {
	default:
		log.Fatalln(v)
		return false
	case *nagiosToJson.Hoststatus:
		res, _ = strconv.Atoi(v.Problem_has_been_acknowledged)
		if res == 1 {
			return true
		}
	case *nagiosToJson.Servicestatus:
		res, _ = strconv.Atoi(v.Problem_has_been_acknowledged)
		if res == 1 {
			return true
		}
	}
	return false
}

// define output in one place, make it clean
func title() {
	// use HOST as host check servicename
	// 	if *mutehost != "" {
	// 		color.Printf("%-10s: %v %-10v %v %v\n", "hostname", "alert", "state", "acked", "output")
	// 	}
	// 	if *muteservice != "" {
	color.Printf("%-10s: %-10v %v %-10v %v %v\n", "hostname", "srvname", "alert", "state", "acked", "output")
	// 	}
}

func output(hostname, servicename, v interface{}) {
	// define quick color/graphic view of output
	var s, m, a string
	if state(v) == 0 {
		s = color.Sprintf("@{g}%-10s", "[OK]")
	} else if state(v) == 1 {
		s = color.Sprintf("@{y}%-10s", "[WARNING]")
	} else if state(v) == 2 {
		s = color.Sprintf("@{r}%-10s", "[CRITICAL]")
	}
	if notifications(v) {
		m = color.Sprintf("@{g}  %v  ", ` `)
	} else {
		m = color.Sprintf("@{r}  %v  ", `Ø`)
	}
	if acknowledged(v) {
		a = color.Sprintf("@{g}  %v  ", `√`)
	} else {
		a = color.Sprintf("@{g}  %v  ", ` `)
	}

	// this just incase if change mind use different syntax
	switch v := v.(type) {
	default:
	case *nagiosToJson.Hoststatus:
		// 		color.Printf("%-10s: %v %v %v %v %v\n", hostname, m, s, a, v.Plugin_output, time.Since(str2time(v.Last_check)))
		color.Printf("%-10s: %-10v %v %v %v %v %v\n", hostname, servicename, m, s, a, v.Plugin_output, time.Since(str2time(v.Last_check)))
	case *nagiosToJson.Servicestatus:
		color.Printf("%-10s: %-10v %v %v %v %v %v\n", hostname, servicename, m, s, a, v.Plugin_output, time.Since(str2time(v.Last_check)))
	}
}

func str2time(str string) time.Time {
	t, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		log.Fatalln(err)
	}
	return time.Unix(t, 0)
}
