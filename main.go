package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/digitalocean/godo"
)

var Config struct {
	APIKey string
	Domain string
	Record string
}

func init() {
	Config.APIKey = os.Getenv("DODDNS_API_KEY")
	Config.Domain = os.Getenv("DODDNS_DOMAIN")
	Config.Record = os.Getenv("DODDNS_RECORD")
}

func GetPublicIP() (string, error) {
	req, err := http.Get("http://ifconfig.me/")
	if err != nil {
		return "", nil
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func updatedRecord(newIp string) *godo.DomainRecordEditRequest {
	return &godo.DomainRecordEditRequest{
		Type: "A",
		Name: Config.Record,
		Data: newIp,
		TTL:  60,
	}
}

func main() {
	publicIP, err := GetPublicIP()
	if err != nil {
		log.Fatalf("Failed to discover public IP")
	}
	log.Printf("Public IP discovered as %s", publicIP)

	client := godo.NewFromToken(Config.APIKey)
	recordEdit := updatedRecord(publicIP)

	dom, resp, err := client.Domains.Records(context.TODO(), Config.Domain, nil)
	log.Println("domain", dom)
	log.Println("response", resp)

	if err != nil {
		log.Fatalf("Failed to fetch domain records: %v", err)
	}

	for _, rec := range dom {
		if rec.Name != Config.Record {
			continue
		}

		if rec.Data != publicIP {
			newRec, resp, err := client.Domains.EditRecord(context.TODO(), Config.Domain, rec.ID, recordEdit)
			log.Println("record", newRec)
			log.Println("response", resp)

			if err != nil {
				log.Fatalf("Failed to update existing record: %v", err)
			}

			log.Printf("Updated existing record to %s", publicIP)
		} else {
			log.Println("Record is up to date!")
		}
		return
	}
	log.Println("Existing record not found, creating a new one!")

	// No record found, create one
	newRec, resp, err := client.Domains.CreateRecord(context.TODO(), Config.Domain, recordEdit)
	log.Println("record", newRec)
	log.Println("response", resp)

	if err != nil {
		log.Fatalf("Failed to create new record: %v", err)
	}

	log.Print("Created a new record")
}
