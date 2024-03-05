package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type BrasilAPIAddress struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

func (b *BrasilAPIAddress) toAddress() Address {
	return Address{
		cep:          b.Cep,
		state:        b.State,
		city:         b.City,
		neighborhood: b.Neighborhood,
		street:       b.Street,
		err:          nil,
	}
}

type ViaCepAddress struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

func (b *ViaCepAddress) toAddress() Address {
	return Address{
		cep:          b.Cep,
		state:        b.Uf,
		city:         b.Localidade,
		neighborhood: b.Bairro,
		street:       b.Logradouro + " - " + b.Complemento,
		err:          nil,
	}
}

type Address struct {
	cep          string
	state        string
	city         string
	neighborhood string
	street       string
	err          error
	origin       string
}

func brasilapi(cep string, c chan<- Address) {
	for { //for para retry
		req, err := http.Get("https://brasilapi.com.br/api/cep/v1/" + cep)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao fazer requisição: %v\n", err)
			continue
		}
		defer req.Body.Close()
		res, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao ler resposta: %v\n", err)
			continue
		}
		var data BrasilAPIAddress
		err = json.Unmarshal(res, &data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao fazer parse da resposta: %v\n", err)
			continue
		}
		c <- data.toAddress()
		break
	}
}

func viacep(cep string, c chan<- Address) {
	for { //for para retry
		req, err := http.Get("http://viacep.com.br/ws/" + cep + "/json/")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao fazer requisição: %v\n", err)
			continue
		}
		defer req.Body.Close()
		res, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao ler resposta: %v\n", err)
			continue
		}
		var data ViaCepAddress
		err = json.Unmarshal(res, &data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao fazer parse da resposta: %v\n", err)
			continue
		}
		c <- data.toAddress()
		break
	}
}

func main() {
	cViaCep := make(chan Address)
	cBrasilapi := make(chan Address)
	cep := "60541646"
	go viacep(cep, cViaCep)
	go brasilapi(cep, cBrasilapi)

	select {
	case address := <-cViaCep:
		fmt.Printf("Endereço recebido da ViaCEP: %s, %s, %s, %s - %s \n", address.street, address.cep, address.neighborhood, address.city, address.state)

	case address := <-cBrasilapi:
		fmt.Printf("Endereço recebido da BrasilAPI: %s, %s, %s, %s - %s \n", address.street, address.cep, address.neighborhood, address.city, address.state)

	case <-time.After(time.Second):
		println("timeout error")
	}
}
