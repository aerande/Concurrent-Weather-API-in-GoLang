
package main  
import (
    "encoding/json"
    "net/http"
    "strings"
    "time"
)

func main() {  

    mw := multiWeatherProvider{
        openWeatherMap{},
        weatherUnderground{},
    }
   
   http.HandleFunc("/hello",hello)


   http.HandleFunc("/weather/",func(w http.ResponseWriter,r *http.Request) {
     begin := time.Now()
    city := strings.SplitN(r.URL.Path,"/",3)[2]
    temp,err := mw.tempreture(city)
    if(err !=nil) {
        http.Error(w,err.Error(),http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type","application/json; charset=utf-8")
    json.NewEncoder(w).Encode(map[string]interface{} {
            "city":city,
            "temp":temp,
            "took":time.Since(begin).String(),
        })

    })
   http.ListenAndServe(":8989",nil)
}

func hello(w http.ResponseWriter,r *http.Request) {

        w.Write([]byte("hello!"))
}
//e7e6cad98af55ad8b6e46cdc5867681c
type openWeatherMap struct{}

 func (w openWeatherMap) tempreture(city string) (float64,error) {

    resp,err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=e7e6cad98af55ad8b6e46cdc5867681c&q=" + city)

        if err != nil {
            return 0.0,err
        }

        defer resp.Body.Close()

        var d  struct {
            Main struct {
                   Kelvin float64 `json:"temp"`
               } `json:"main"`
        }

        if err := json.NewDecoder(resp.Body).Decode(&d); err!=nil {
            return 0,err
        }

        return d.Main.Kelvin,nil
}

type weatherUnderground struct{}

 func (w weatherUnderground) tempreture(city string) (float64,error) {

    resp,err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=e7e6cad98af55ad8b6e46cdc5867681c&q=" + city)

        if err != nil {
            return 0.0,err
        }

        defer resp.Body.Close()

        var d  struct {
            Main struct {
                   Kelvin float64 `json:"temp"`
               } `json:"main"`
        }

        if err := json.NewDecoder(resp.Body).Decode(&d); err!=nil {
            return 0,err
        }

        return d.Main.Kelvin,nil
}

type weatherData struct {
    Name string `json:"name"`
    Main struct {
        Kelvin float64 `json:"temp"`
    } `json:"main"`
}

type weatherProvider interface {
    tempreture(city string)(float64,error)
}


func tempreture(city string,providers ...weatherProvider)(float64,error) {
        sum:=0.0
    for _,provider := range providers {
        k,err := provider.tempreture(city)

        if err !=nil {
            return 0,err
        }

        sum+=k
    }

    return sum/(float64(len(providers))),nil
}


type multiWeatherProvider []weatherProvider

func (mw multiWeatherProvider) tempreture(city string) (float64,error) {

    temps:= make(chan float64,len(mw))
    errs:= make(chan error,len(mw))


    sum :=0.0

    for _,provider:= range mw {

            go func(p weatherProvider) {
            k,err := provider.tempreture(city)

            if err!=nil {
                errs <- err
                return  
            }
            temps <-k
            }(provider)
    }

    for i:=0;i<len(mw);i++ {
        select {
        case tem:= <-temps:
            sum +=tem
        case err:= <-errs:
            return 0,err
        }
    }
     return sum/(float64(len(mw))),nil
    }
