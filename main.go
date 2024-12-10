package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	httpSwagger "github.com/swaggo/http-swagger"
)

type RetornoAPI struct {
	Cep              string `json:"cep"`
	State            string `json:"state"`
	City             string `json:"city"`
	Neighborhood     string `json:"neighborhood"`
	Street           string `json:"street"`
	Service          string `json:"service"`
	tempo_requisicao string `json:"tempo_requisicao"`
}

var Cep string = ""

func main() {

	r := chi.NewRouter()
	r.Route("/CEP", func(r chi.Router) {
		r.Get("/{cep}", GetCEP)
	})

	r.Get("/docs/*", httpSwagger.Handler(httpSwagger.URL("http://localhost:8000/docs/doc.json")))
	http.ListenAndServe(":8000", r)
}

func GetCEP(w http.ResponseWriter, r *http.Request) {
	Cep = chi.URLParam(r, "cep")
	//wg := sync.WaitGroup{}
	canal_api_1 := make(chan string)
	canal_api_2 := make(chan string)

	go func() {
		canal_api_1 <- GetCEP_viaCEP(Cep)
	}()
	go func() {
		canal_api_2 <- GetCEP_brasilApi(Cep)
	}()

	// Usar select para esperar pelas respostas das APIs
	select {
	case msg := <-canal_api_1: // resposta da Brasil API
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(msg)

	case msg := <-canal_api_2: // resposta da ViaCEP API
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(msg)

		// case <-time.After(time.Second * 3): // Timeout após 3 segundos
		// 	log.Println("timeout")
	}
}

func GetCEP_brasilApi(cep string) string {
	start := time.Now()
	resp, err := http.Get("https://brasilapi.com.br/api/cep/v1/" + cep)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var retorno RetornoAPI
	duration := time.Since(start)

	err = json.Unmarshal(body, &retorno)
	if err != nil {
		log.Fatalf("Erro ao deserializar o JSON: %v", err)
	}

	jsonData := make(map[string]interface{})
	jsonData["tempo_requisicao"] = duration
	jsonData["api_acatada"] = "brasil_api"
	jsonData["Cep"] = retorno.Cep
	jsonData["City"] = retorno.City
	jsonData["Street"] = retorno.Street
	jsonData["Service"] = retorno.Service

	json, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		log.Fatalf("Erro ao serializar o JSON: %v", err)
	}

	return string(json)
}

func GetCEP_viaCEP(cep string) string {
	start := time.Now()
	resp, err := http.Get("http://viacep.com.br/ws/" + cep + "/json/")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var retorno RetornoAPI
	duration := time.Since(start)

	err = json.Unmarshal(body, &retorno)
	if err != nil {
		log.Fatalf("Erro ao deserializar o JSON: %v", err)
	}

	jsonData := make(map[string]interface{})
	jsonData["tempo_requisicao"] = duration
	jsonData["api_acatada"] = "via_cep"
	jsonData["Cep"] = retorno.Cep
	jsonData["City"] = retorno.City
	jsonData["Street"] = retorno.Street
	jsonData["Service"] = retorno.Service

	json, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		log.Fatalf("Erro ao serializar o JSON: %v", err)
	}

	return string(json)

}