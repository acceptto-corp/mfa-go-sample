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
    u := User { Email: "a.karimi.k@gmail.com", MfaType: 1 }

    m := martini.Classic()

    m.Get("/acceptto-callback", func() string {
        return "Callback page"
    });
    m.Get("/enable", func(ctx martini.Context) string {
        AccepttoEnableMfaUser(u, ctx)
        return "Ok. Enabled."
    });
    m.Get("/disable", func(ctx martini.Context) string {
        AccepttoDisableMfaUser(u, ctx)
        return "Ok. Disabled."
    });
    m.Get("/auth", func(ctx martini.Context) string {
        AccepttoAuthenticateMfaUser(u, ctx)
        // redirect to https://mfa.acceptto.com/mfa/index?channel=channel_you_got_from_step_3&callback_url=http://your_domain.com/auth/mfa_check
        return 
    }    
    
    m.Run()
}

func AccepttoEnableMfaUser(user User, ctx martini.Context) {
	log.Printf("acceptto.go:AccepttoDisableMfaUser:%+v", user)
	approved, err := AcceptoSendRequest(user.Email, "Confirm Acceptto Authentication")
	if err != nil {
		ctx.Map( fmt.Errorf("Error: %s", err))
		return
	}
	if approved {
		user.MfaType=1
		// TODO
		//user.Update(c)
		ctx.Map(user)
		return
	} else {
		ctx.Map( fmt.Errorf("User %s Not Approved", user.Email))
		return
	}
}

func AccepttoDisableMfaUser(user User, ctx martini.Context){
	log.Printf("acceptto.go:AccepttoDisableMfaUser:%+v", user)
	approved, err := AcceptoSendRequest(user.Email, "Confirm Acceptto Authentication Removal")
	if err != nil {
		ctx.Map( fmt.Errorf("Error: %s", err))
		return
	}
	if approved {
		return
	} else {
		ctx.Map( fmt.Errorf("User %s Not Approved", user.Email))
		return
	}
}
func AccepttoAuthenticateMfaUser(user User, ctx martini.Context) (err error, channel string) {
	log.Printf("acceptto.go:AccepttoAuthenticateMfaUser:%+v", user)
	approved, err, channel := AcceptoSendRequest(user.Email, "Login to Your Awesome Service")
	if err != nil {
		ctx.Map(fmt.Errorf("Error: %s", err))
		return err, ""
	}
	if approved {
		return nil, channel
	} else {
		ctx.Map(fmt.Errorf("User %s Not Approved", user.Email))
		return fmt.Errorf("User %s Not Approved", user.Email), ""
	}
}
func AcceptoSendRequest(email string, message string) (mfa_result bool, err error, channel string){
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
	cannel = data.Channel
	
	for i:=0;i<30;i++ {
		time.Sleep(2 * time.Second)
		resp_status, err := http.Get("https://mfa.acceptto.com/api/v9/check?channel="+data.Channel+"&email="+email)
		if err == nil {
			body_status, err := ioutil.ReadAll(resp_status.Body)
			if err == nil {
				log.Printf("acceptto.go:AcceptoSendRequest:received:%s", body_status)
				break;
				
				var accepttoStatus AccepttoStatus
				json.Unmarshal(body_status, &accepttoStatus)
				if accepttoStatus.Status=="approved"{
					log.Printf("acceptto.go:AcceptoSendRequest:approved:%+v", body_status)
					mfa_result=true;
					break;
				}
				if accepttoStatus.Status=="rejected"{
					log.Printf("acceptto.go:AcceptoSendRequest:rejected:%+v", body_status)
					err=fmt.Errorf("MFA %s Rejected", email)
					break;
				}
			}
		}
			
	}
	if err == nil && mfa_result==false{
		log.Printf("acceptto.go:AcceptoSendRequest:timeout:%+v", data)
		err=fmt.Errorf("MFA %s Timeout", email)
	}
	return
}
