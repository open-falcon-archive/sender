package cron

import (
	"log"
	"strings"
	"time"

	"github.com/open-falcon/sender/g"
	"github.com/open-falcon/sender/model"
	"github.com/open-falcon/sender/proc"
	"github.com/open-falcon/sender/redis"
	"github.com/toolkits/net/httplib"
)

func ConsumeSms() {
	queue := g.Config().Queue.Sms
	for {
		L := redis.PopAllSms(queue)
		if len(L) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendSmsList(L)
	}
}

func SendSmsList(L []*model.Sms) {
	for _, sms := range L {
		SmsWorkerChan <- 1
		go SendSms(sms)
	}
}

func SendSms(sms *model.Sms) {
	defer func() {
		<-SmsWorkerChan
	}()

	if g.Config().Debug {
		log.Println("==sms==>>>>", sms)
	}

	tos := strings.Replace(sms.Tos, ";", ",", -1)
	arr := strings.Split(tos, ",")
	for i := 0; i < len(arr); i++ {
		if arr[i] == "" {
			continue
		}
		err := _sms(arr[i], sms.Content)
		if err != nil {
			log.Println(err)
		}
	}
}

func _sms(phone, content string) error {
	cfg := g.Config()
	url := cfg.Api.Sms
	r := httplib.Get(url).SetTimeout(5*time.Second, 2*time.Minute)
	r.Param("PHONE", phone)
	r.Param("CONTENT", content)
	r.Param("UID", cfg.Sms.UID)
	r.Param("PWD", cfg.Sms.PWD)
	r.Param("TYPE", cfg.Sms.TYPE)
	r.Param("MSGID", cfg.Sms.MSGID)
	_, err := r.String()
	if err != nil {
		log.Println(err)
		return err
	}

	proc.IncreSmsCount()
	return nil
}
