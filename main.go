package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Estrutura genérica para representar a resposta de ambas as APIs
type Endereco struct {
	Cep        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"` // ViaCEP usa isso
	Uf         string `json:"uf"`
	Estado     string `json:"state"` // BrasilAPI usa isso
	Cidade     string `json:"city"`  // BrasilAPI usa isso
	ApiUsada   string
}

// Função para buscar no BrasilAPI
func fetchFromBrasilAPI(ctx context.Context, cep string, ch chan<- Endereco) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var endereco Endereco
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &endereco); err != nil {
		return
	}
	endereco.ApiUsada = "BrasilAPI"
	ch <- endereco
}

// Função para buscar no ViaCEP
func fetchFromViaCEP(ctx context.Context, cep string, ch chan<- Endereco) {
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var endereco Endereco
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &endereco); err != nil {
		return
	}
	endereco.ApiUsada = "ViaCEP"
	ch <- endereco
}

func main() {
	//cep := "01153000" // Exemplo de CEP inválido defnido para teste
	cep := "54325251" // Exemplo de CEP válido
	resultCh := make(chan Endereco, 2)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go fetchFromBrasilAPI(ctx, cep, resultCh)
	go fetchFromViaCEP(ctx, cep, resultCh)

	select {
	case res := <-resultCh:
		fmt.Println("✅ Resposta recebida da API:", res.ApiUsada)
		fmt.Println("📍 CEP:", res.Cep)
		fmt.Println("📍 Logradouro:", res.Logradouro)
		fmt.Println("📍 Bairro:", res.Bairro)
		fmt.Println("📍 Cidade:", res.Localidade+res.Cidade) // uma das duas estará vazia
		fmt.Println("📍 Estado:", res.Uf+res.Estado)         // uma das duas estará vazia
	case <-ctx.Done():
		fmt.Println("⏱️ Erro: timeout após 1 segundo.")
	}
}
