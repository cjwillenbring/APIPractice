//remember to set GOOGLE_APPLICATION_CREDENTIALS=[PATH] for each new session
/** curl requests
GET: curl -X GET localhost:8000/Trips/{documentid}
POST: curl -X POST localhost:8000/Trips/{documentid}
PATCH: curl -X PATCH localhost:8000/Trips/{documentid}
DELETE: curl -X "DELETE" localhost:8000/Trips/{documentid}
*/

/** http body of post and patch requests
Example Request:
{
  "Name": "tripster",
  "Location": {
	   "City":"Overland Park", "Place":"Running Trails", "State":"Kansas", "ContactInfo":"(913)401-9930"
  },
  "TimeFrame": {
    	"StartTime": "9:30", "EndTime": "10:30"
  },
  "Members": [
    	{"Name": "Cole","IsComing":true,"Username":"LittleDog"},
    	{"Name": "Max","IsComing":true,"Username":"BigDawg"},
    	{"Name": "Mahmood","IsComing":false}
	]
}
*/

package main

import (
  "context"
  "log"
	"fmt"
  "net/http"
  "encoding/json"
  "io/ioutil"
  "github.com/gorilla/mux"
  "cloud.google.com/go/firestore"
)

type GoDAO struct {
  Context context.Context
  Client *firestore.Client
  Document string
}

type Trip struct {
  Location Location `firestore:"location,omitempty"`
  TimeFrame TimeFrame `firestore:"timeFrame,omitempty"`
  Name string `firestore:"name,omitempty"`
  Members []Member `firestore:"members,omitempty"`
}

type Location struct {
  City string `firestore:"city,omitempty"`
  ContactInfo string `firestore:"contactInfo,omitempty"`
  Place string `firestore:"place,omitempty"`
  State string `firestore:"state,omitempty"`
}

type TimeFrame struct {
  StartTime string `firestore:"startTime,omitempty"`
  EndTime string `firestore:"endTime,omitempty"`
}

type Member struct {
  IsComing bool `firestore:"isComing,omitempty"`
  Name string `firestore:"name,omitempty"`
  Username string `firestore:"username,omitempty"`
}

func main()  {
  rout := mux.NewRouter()
  rout.HandleFunc("/Trips/{documentid}", MakeHttpDBHandler(readTrip)).Methods("GET") //Read trip info
  rout.HandleFunc("/Trips/{documentid}", MakeHttpDBHandler(createTrip)).Methods("POST") //Create new trip document
  rout.HandleFunc("/Trips/{documentid}", MakeHttpDBHandler(updateTrip)).Methods("PATCH") //Update existing trip
  rout.HandleFunc("/Trips/{documentid}", MakeHttpDBHandler(deleteTrip)).Methods("DELETE") //Delete trip document
  log.Fatal(http.ListenAndServe(":8000", rout)) //listen and route http requests at port 8000
}

func MakeHttpDBHandler(fn func (http.ResponseWriter, *http.Request, GoDAO)) http.HandlerFunc {
  return func (w http.ResponseWriter, r *http.Request)  {
    ctx := context.Background()
    projectID := "golangpracticeproject"
    client, err := firestore.NewClient(ctx, projectID)
    if err != nil {
  		log.Fatalf("Failed to create client: %v", err) //handle errors
  	}
    defer client.Close()
    vars := mux.Vars(r)
    dao := GoDAO { Context: ctx, Client: client, Document: vars["documentid"]}
    fn(w,r,dao)
  }
}

func readTrip(w http.ResponseWriter, r *http.Request, GoDAO GoDAO) {
  doc, err := GoDAO.Client.Collection("Trips").Doc(GoDAO.Document).Get(GoDAO.Context)
  if err != nil {
    log.Fatalf("Failed to create client: %v", err)
  }
  m := doc.Data()
  fmt.Fprintf(w,"Data from " + GoDAO.Document + "\nDocument Data: ")
  jsonData, err := json.Marshal(m)
  if err != nil {
    log.Fatalf("Failed converting to json")
  }
  fmt.Fprintf(w,string(jsonData))
}

func createTrip(w http.ResponseWriter, r *http.Request, GoDAO GoDAO) {
  var newTrip Trip

  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    log.Printf("An error has occurred: %s", err)
  }

  error := json.Unmarshal(body, &newTrip)
  if error != nil {
    log.Printf("An error has occurred: %s", error)
  }

  fmt.Fprintf(w,"%#v",newTrip)

  _, erg := GoDAO.Client.Collection("Trips").Doc(GoDAO.Document).Set(GoDAO.Context, newTrip)
  if erg != nil {
    log.Printf("An error has occurred: %s", erg)
  }
  fmt.Fprintf(w,"Document " + GoDAO.Document + " was created")
}

func updateTrip(w http.ResponseWriter, r *http.Request, GoDAO GoDAO) {
  var tripToUpdate map[string]interface{}

  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    log.Printf("An error has occurred: %s", err)
  }

  error := json.Unmarshal(body, &tripToUpdate)
  if error != nil {
    log.Printf("An error has occurred: %s", error)
  }

  _, erg := GoDAO.Client.Collection("Trips").Doc(GoDAO.Document).Set(GoDAO.Context, tripToUpdate, firestore.MergeAll)
  if erg != nil {
    log.Printf("An error has occurred: %s", erg)
  }
  fmt.Fprintf(w,"Document " + GoDAO.Document + " was updated")
}

func deleteTrip(w http.ResponseWriter, r *http.Request, GoDAO GoDAO) {
  _, err := GoDAO.Client.Collection("Trips").Doc(GoDAO.Document).Delete(GoDAO.Context)
  if err != nil {
    log.Printf("An error has occurred: %s", err)
  }
  fmt.Fprintf(w,"Document " + GoDAO.Document + " was deleted")
}
