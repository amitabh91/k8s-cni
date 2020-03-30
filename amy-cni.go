package main
import (
    "fmt"
    "encoding/json"
//  "os"
    "os/exec"
    "net/http"
    "io/ioutil"
    "log"
    "strings"
)


type InterfaceDetails struct {
NetworkName string `json: "networkname"`
InterfaceName string `json: "interfacename"`
IP  string `json: "ip"`
ContainerID string `json: "containerid"`
VIP string `json: "vip"`
Namespace string `json: "namespace"`
}

func main() {

http.HandleFunc("/v1/vip/add_interface", addInterface)
http.HandleFunc("/v1/vip/del_interface", deleteInterface)

log.Fatal(http.ListenAndServe(":9090", nil))

}

const(
cniToolPath="/host/root/cnitool"
)

var (
       staticPodRoute string
)
func addInterface(w http.ResponseWriter, r *http.Request) {
    reqBody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Println("error while reading the request body", err)
       w.WriteHeader(400)
        return
    }

    var networkDetails InterfaceDetails
 
    err = json.Unmarshal(reqBody, &networkDetails)
    if err != nil {
        log.Println("error while unmarshalling the request", err)
       w.WriteHeader(400)
        return
    }

    pid:= processNetworkDetails(networkDetails)
    if pid== ""{
log.Println("pid is undefined")
w.WriteHeader(500)
return
    }
   
    vip:= networkDetails.VIP
    ip := exec.Command("/bin/bash", "-c", "sed 's/200.200.200.20/"+vip+"/g' /host/root/conf/secondary.conf > /host/root/network/secondary.conf")
    ip.Output()
   
    pid = strings.TrimSpace(pid)

    // add interface

cmd:= exec.Command("/bin/bash","-c","NETCONFPATH=/host/root/network/ "+cniToolPath+" add "+networkDetails.NetworkName+" /proc/"+pid+"/ns/net "+networkDetails.InterfaceName+" "+strings.TrimSpace(staticPodRoute))
stdout, err := cmd.Output()
    if err != nil {
       //log.Println(err.Error())
       fmt.Println("error while adding interface", err.Error())
       w.WriteHeader(500)
       return
    }

    fmt.Println("interface added sucessfully for", string(stdout))
    w.WriteHeader(200)
    json.NewEncoder(w).Encode(string(stdout))
}

func deleteInterface(w http.ResponseWriter, r *http.Request) {
    reqBody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Println("error while reading the request body", err)
       w.WriteHeader(400)
        return
    }

    var networkDetails InterfaceDetails

    err = json.Unmarshal(reqBody, &networkDetails)
    if err != nil {
        log.Println("error while unmarshalling the request", err)
       w.WriteHeader(400)
return
    }

    pid:= processNetworkDetails(networkDetails)
    if pid == "" {
   log.Println("pid is undefined")
            w.WriteHeader(500)
   return
    }

    pid = strings.TrimSpace(pid)

        //Del interface

cmd:= exec.Command("/bin/bash", "-c", "NETCONFPATH=/host/root/network/ "+cniToolPath+" del "+networkDetails.NetworkName+" /proc/"+pid+"/ns/net "+networkDetails.InterfaceName+" "+strings.TrimSpace(staticPodRoute))

stdout, err := cmd.Output()
    if err != nil {
       //log.Println(err.Error())
       fmt.Println(fmt.Sprint(err) + ": " + err.Error())
       w.WriteHeader(500)
       return
    }

    fmt.Println("Interface deleted successfully for", string(stdout))
    w.WriteHeader(200)
    json.NewEncoder(w).Encode(string(stdout))
}

func processNetworkDetails(networkDetails InterfaceDetails) string {
     cmd := exec.Command("/bin/bash" , "-c", "docker inspect --format {{.State.Pid}} "+networkDetails.ContainerID)
     stdout, err := cmd.Output()
     if err != nil {
         //log.Println(err.Error())
fmt.Println(fmt.Sprint(err) + ": " + err.Error())
         return ""
      }

      pid:= string(stdout)

     cmd = exec.Command("/bin/bash" , "-c", "route -n |grep "+networkDetails.IP+" | awk '{print $8}' | cut -c 5-")
     stdout, err = cmd.Output()
     if err != nil {
         //log.Println(err.Error())
fmt.Println(fmt.Sprint(err) + ": " + err.Error())
         return ""
      }

      staticPodRoute= string(stdout)
      return pid
}
