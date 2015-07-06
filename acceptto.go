package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"github.com/codegangsta/martini"
	"net/http"
	"time"
    "io/ioutil"
)

const uid = "cd279a7c988afeabbea999c66b294be4c99d466db9d5032ba00150124f0b00a1"
const secret = "6bee88f143ee21b09b7054e800eaba918fb45890151088819e040ebc83f743de"
const email = "a.karimi.k@gmail.com"

// Sample user
type User struct {
    Email   string
    MfaType int
}

type AccepttoRequest struct {
    Channel string  `json:"channel,omitempty"`
}

type AccepttoStatus struct {
    Status  string  `json:"status,omitempty"`
}

func main() {
    u := User { Email: email }
    m := martini.Classic()

    m.Get("/enable", func(ctx martini.Context) string {
        _, err := AccepttoEnableMfaUser(u, ctx)
        if err != nil {
            return fmt.Sprintf("Error: %q", err)
        }
        return "Ok. Enabled."
    });
    m.Get("/disable", func(ctx martini.Context) string {
        _, err := AccepttoDisableMfaUser(u, ctx)
        if err != nil {
            return fmt.Sprintf("Error: %q", err)
        }
        return "Ok. Disabled."
    });
    m.Get("/auth", func(ctx martini.Context) string {
        _, err := AccepttoAuthenticateMfaUser(u, ctx)
        if err != nil {
            return fmt.Sprintf("Error: %q", err)
        }
        return "Ok. authorized"
    })
    
    m.Run()
}

func AccepttoEnableMfaUser(user User, ctx martini.Context) (channel string, err error) {
	log.Printf("acceptto.go:AccepttoDisableMfaUser:%+v", user)
	approved, channel, err := AcceptoSendRequest(user.Email, "Confirm Acceptto Authentication")
	if err != nil {
		return
	}
	if approved {
		return
	} else {
	    err = fmt.Errorf("User %s Not Approved", user.Email)
		return 
	}
}

func AccepttoDisableMfaUser(user User, ctx martini.Context) (channel string, err error) {
	log.Printf("acceptto.go:AccepttoDisableMfaUser:%+v", user)
	approved, channel, err := AcceptoSendRequest(user.Email, "Confirm Acceptto Authentication Removal")
	if err != nil {
		return
	}
	if approved {
		return
	} else {
	    err = fmt.Errorf("User %s Not Approved", user.Email)
		return 
	}
}
func AccepttoAuthenticateMfaUser(user User, ctx martini.Context) (channel string, err error) {
	log.Printf("acceptto.go:AccepttoAuthenticateMfaUser:%+v", user)
	approved, channel, err := AcceptoSendRequest(user.Email, "Login to Your Awesome Service")
	if err != nil {
		return
	}
	if approved {
		return
	} else {
	    err = fmt.Errorf("User %s Not Approved", user.Email)
		return 
	}
}
func AcceptoSendRequest(email string, message string) (mfa_result bool, channel string, err error){
    // Better solution is to redirect to get channel from acceptto API, store it somewhere (like session) and then redirect user to the following page
    // redirect to https://mfa.acceptto.com/mfa/index?channel=channel_you_got_from_step_3&callback_url=http://your_domain.com/auth/mfa_check
	
	mfa_result = false
	message_encoded := url.QueryEscape(message)
	log.Printf("acceptto.go:AcceptoSendRequest:%+v", email)
	resp, err := http.Get("https://mfa.acceptto.com/api/v9/authenticate?timeout=60&message="+message_encoded+"&type=Login&email="+email+"&uid="+uid+"&secret="+secret)
	
	if err != nil {
		log.Printf("acceptto.go:AcceptoSendRequest:error:respond%+v", err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("acceptto.go:AcceptoSendRequest:error:body%+v", err)
		return
	}
	var data AccepttoRequest
	json.Unmarshal(body, &data)
	log.Printf("acceptto request received: %+v", data)
	channel = data.Channel
	
	final_err := err
	
	for i:=0;i<30;i++ {
		time.Sleep(2 * time.Second)
		resp_status, err := http.Get("https://mfa.acceptto.com/api/v9/check?channel="+data.Channel+"&email="+email)
		if err == nil {
			body_status, err := ioutil.ReadAll(resp_status.Body)
			if err == nil {
				log.Printf("acceptto.go:AcceptoSendRequest:received:%s", body_status)
				var accepttoStatus AccepttoStatus
				json.Unmarshal(body_status, &accepttoStatus)
				if accepttoStatus.Status=="approved"{
					log.Printf("acceptto.go:AcceptoSendRequest:approved:%+v", body_status)
					mfa_result=true;
					break;
				}
				if accepttoStatus.Status=="rejected"{
					log.Printf("acceptto.go:AcceptoSendRequest:rejected:%+v", body_status)
					final_err = fmt.Errorf("MFA %s Rejected", email)
					break;
				}
				if accepttoStatus.Status=="expired"{
					log.Printf("acceptto.go:AcceptoSendRequest:expired:%+v", body_status)
					break;
				}
			}
		}
	}
	
    if final_err == nil && mfa_result == false {
		log.Printf("acceptto.go:AcceptoSendRequest:timeout:%+v", data)
		err=fmt.Errorf("MFA %s Timeout", email)
	}
	return
}
